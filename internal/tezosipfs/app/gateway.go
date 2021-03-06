package app

import (
	"bytes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/cache"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/common"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/db"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/network"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/swarm"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

type Gateway struct {
	log            *logrus.Entry
	net            network.NetworkInterface
	cache          cache.Cache
	port           int
	swarm          *swarm.Swarm
	accessTokens   []config.AccessTokens
	uploadTokens   []config.AccessTokens
	l              *sync.Mutex
	db             *db.StormDB
	c              *config.Config
	pendingUploads map[string]*PendingUpload
}

func NewGateway(c *config.Config, net network.NetworkInterface, l *logrus.Entry, s *swarm.Swarm, db *db.StormDB) *Gateway {
	if !c.GatewayEnabled {
		l.Info("HTTP Gateway disabled")
		return nil
	}
	g := Gateway{}
	g.net = net
	g.swarm = s
	g.db = db
	g.c = c
	g.pendingUploads = map[string]*PendingUpload{}
	g.l = &sync.Mutex{}
	g.log = l.WithField("source", "gateway")
	g.port = c.Gateway.Server.Port
	if c.Gateway.Storage.S3.Bucket != "" {
		g.log.Info("Using S3 as storage backend")
		cache := cache.NewS3Cache(c, l)
		g.cache = cache
	}
	if g.port <= 1 {
		g.log.Panic("Invalid Gateway port")
	}

	if g.cache == nil {
		g.log.Warn("Running gateway without storage cache!")
	}
	g.uploadTokens = []config.AccessTokens{}
	g.accessTokens = []config.AccessTokens{}
	go g.watchConfig(c)
	go g.autocache()
	return &g
}

func (g *Gateway) Run() {
	go g.listen()
	g.log.Info("Starting gateway on port :" + strconv.Itoa(g.port))
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	if len(g.c.Gateway.CORS.AllowedDomains) >= 1 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     g.c.Gateway.CORS.AllowedDomains,
			AllowMethods:     []string{"GET", "OPTIONS"},
			AllowHeaders:     []string{"Origin"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			// max age of prefilght cache
			MaxAge: 12 * time.Hour,
		}))
	}

	r.GET("/ipfs/:cid", g.ipfsRoute)
	r.POST("/upload", g.uploadRoute)
	r.POST("/upload/once", g.onceUploadRoute)
	r.POST("/upload/store_and_cache", g.oncStoreAndCachedUploadRoute)
	r.POST("/upload/threshold", g.customThreshold)
	r.GET("/network", g.networkRoute)
	r.Run("0.0.0.0:" + strconv.Itoa(g.port))
}

func (g *Gateway) ipfsRoute(c *gin.Context) {

	if g.checkAccessToken(c) {
		return
	}

	cid := c.Param("cid")
	headers := map[string]string{
		"Cache-Control": "max-age=86400", // cache for one day, ipfs content never changes
	}
	if len(cid) <= 12 || len(cid) >= 64 {
		c.String(500, "invalid cid")
		return
	}

	if g.db.IsBlocked(cid) {
		c.String(404, "not found")
		return
	}
	l, reader, err := g.cache.GetFile(cid)
	if err == nil {
		g.log.WithField("cid", cid).Trace("Cache hit")
		buf, _ := ioutil.ReadAll(reader)
		c.DataFromReader(200, l, getType(buf), bytes.NewReader(buf), headers)
		return
	}
	reader, err = g.net.GetFile(c, cid)
	if err != nil {
		c.String(404, ":(")
		return
	}
	g.log.WithField("cid", cid).Trace("Found via Network")
	buf, _ := ioutil.ReadAll(reader)
	g.cache.StoreFile(cid, bytes.NewReader(buf))
	// save a record in db for future use
	cacheEntry := common.Cache{
		Created: time.Now(),
		Cid:     cid,
		From:    "gateway",
		Status:  "cached",
		Size:    int64(len(buf)),
	}
	g.db.SaveCache(&cacheEntry)
	c.DataFromReader(200, int64(len(buf)), getType(buf), bytes.NewReader(buf), headers)
}

func (g *Gateway) networkRoute(c *gin.Context) {
	if g.checkAccessToken(c) {
		return
	}
	res := g.getNetwork()
	c.JSON(200, res)
}

func (g *Gateway) getNetwork() UploadResponse {
	// abuse UploadResponse, almost same schema
	res := UploadResponse{}
	res.CacheNodes = []CacheNode{}
	res.StorageNodes = []StorageNode{}
	caches := g.swarm.CacheForUs()
	stores := g.swarm.PinForUs()
	for _, c := range caches {
		res.CacheNodes = append(res.CacheNodes, CacheNode{
			Name:         c.Name,
			Organization: c.Organization,
			Contact:      c.Contact,
			Comment:      c.Comment,
			PeerId:       c.PeerId,
		})
	}
	for _, c := range stores {
		res.StorageNodes = append(res.StorageNodes, StorageNode{
			Name:         c.Name,
			Organization: c.Organization,
			Contact:      c.Contact,
			Comment:      c.Comment,
			PeerId:       c.PeerId,
		})
	}
	nc := len(caches)
	ns := len(stores)
	res.NumberCaches = &nc
	res.NumberStores = &ns
	return res
}

func (g *Gateway) uploadRoute(c *gin.Context) {
	if g.checkUploadToken(c) {
		return
	}
	cid, done := g.storeFile(c)
	if done {
		return
	}
	res := UploadResponse{
		Cid: cid,
	}
	c.JSON(200, res)
}

/*
 * Returns after timeout or after at least one node
 * we trust has confirmed the pin
 */
func (g *Gateway) onceUploadRoute(c *gin.Context) {

	check, ticker, done2 := g.prepareGuaranteedUpload(c)
	if done2 {
		return
	}

outer:
	for {
		select {
		case <-ticker.C:
			// got timeout
			ticker.Stop()
			check.lock.Lock()
			g.log.WithField("cid", check.res.Cid).Warn("Once has reached timeout")
			check.res.Status = "Timeout"
			check.lock.Unlock()
			break outer

		case <-check.Notify:
			check.lock.Lock()
			if *check.res.NumberStored >= 1 {
				check.res.Status = "Success"
				check.lock.Unlock()
				break outer
			} else {
				check.lock.Unlock()
			}
		}
	}

	check.lock.Lock()
	defer check.lock.Unlock()
	c.JSON(200, check.res)
	delete(g.pendingUploads, check.res.Cid)
}

/*
 * Returns after timeout or after at least one node
 * we trust has cached the pin
 */
func (g *Gateway) oncStoreAndCachedUploadRoute(c *gin.Context) {

	check, ticker, done2 := g.prepareGuaranteedUpload(c)
	if done2 {
		return
	}

outer:
	for {
		select {
		case <-ticker.C:
			// got timeout
			ticker.Stop()
			check.lock.Lock()
			g.log.WithField("cid", check.res.Cid).Warn("StoreAndCached has reached timeout")
			check.res.Status = "Timeout"
			check.lock.Unlock()
			break outer

		case <-check.Notify:
			check.lock.Lock()
			if *check.res.NumberStored >= 1 && *check.res.NumberCached >= 1 {
				check.res.Status = "Success"
				check.lock.Unlock()
				break outer
			}
			check.lock.Unlock()

		}
	}

	check.lock.Lock()
	defer check.lock.Unlock()
	c.JSON(200, check.res)
	delete(g.pendingUploads, check.res.Cid)
}

/*
 * Returns after timeout or custom guarantees
 */
func (g *Gateway) customThreshold(c *gin.Context) {

	check, ticker, done2 := g.prepareGuaranteedUpload(c)
	if done2 {
		return
	}

	MustStore, _ := strconv.Atoi(c.PostForm("store"))
	MustCache, _ := strconv.Atoi(c.PostForm("cache"))

outer:
	for {
		select {
		case <-ticker.C:
			// got timeout
			ticker.Stop()
			check.lock.Lock()
			g.log.WithField("custom_store", MustStore).
				WithField("custom_cache", MustCache).
				WithField("cid", check.res.Cid).Warn("Custom Store has reached timeout")
			check.res.Status = "Timeout"
			check.lock.Unlock()
			break outer

		case <-check.Notify:
			check.lock.Lock()
			if *check.res.NumberStored >= MustStore && *check.res.NumberCached >= MustCache {
				check.res.Status = "Success"
				check.lock.Unlock()
				break outer
			}
			check.lock.Unlock()

		}
	}

	check.lock.Lock()
	defer check.lock.Unlock()
	c.JSON(200, check.res)
	delete(g.pendingUploads, check.res.Cid)
}

func (g *Gateway) prepareGuaranteedUpload(c *gin.Context) (*PendingUpload, *time.Ticker, bool) {

	timeout_duration := 5 * time.Second
	CustomTimeout, _ := strconv.Atoi(c.PostForm("timeout"))
	if CustomTimeout >= 5 {
		timeout_duration = time.Duration(CustomTimeout) * time.Second
	}

	if g.checkAccessToken(c) {
		return nil, nil, true
	}

	net := g.getNetwork()
	if *net.NumberStores == 0 {
		c.String(500, "Not enough Nodes configured!")
		return nil, nil, true
	}
	net.NumberCached = intptr(0)
	net.NumberStored = intptr(0)

	notify := make(chan struct{})
	check := &PendingUpload{
		lock:   &sync.Mutex{},
		res:    &net,
		Notify: notify,
	}

	cid, done := g.storeFile(c)
	if done {
		return nil, nil, true
	}

	check.res.Cid = cid
	g.pendingUploads[cid] = check
	ticker := time.NewTicker(timeout_duration)
	return check, ticker, false
}
