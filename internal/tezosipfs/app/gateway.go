package app

import (
	"bytes"
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/cache"
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

func (g *Gateway) watchConfig(c *config.Config) {
	ch := c.GetUpdates()
	for {
		n := <-ch
		g.l.Lock()
		g.log.Info("updating config")
		g.accessTokens = n.Gateway.Server.AccessTokens
		g.uploadTokens = n.Gateway.Server.UploadToken
		g.l.Unlock()
	}
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
	r.GET("/network", g.networkRoute)
	r.Run("0.0.0.0:" + strconv.Itoa(g.port))
}

func (g *Gateway) autocache() {
	ch := g.net.Subscribe()
	for {
		msg := <-ch
		if msg.Kind == "new_object" && g.swarm.CacheFor(msg.From) {
			g.log.WithField("source", msg.From).WithField("cid", string(msg.Data)).Trace("pin request")
			if g.swarm.CacheFor(msg.From) {
				g.log.WithField("cid", string(msg.Data)).WithField("origin", msg.From).Info("Auto-cache")
				go g.cacheFile(string(msg.Data))
			}
		}
	}
}

func (g *Gateway) cacheFile(cid string) {
	reader, err := g.net.GetFile(context.Background(), cid)
	if err != nil {
		g.log.Error(err)
	}
	buf, _ := ioutil.ReadAll(reader)
	if g.cache != nil {
		g.cache.StoreFile(cid, bytes.NewReader(buf))
		g.log.WithField("cid", cid).Trace("Stored in cache")
		g.broadcastCache(cid)
	} else {
		g.log.WithField("cid", cid).Error("got cache request, but have no cache configured...")
	}
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

func (g *Gateway) listen() {
	ch := g.net.Subscribe()
	for {
		msg := <-ch
		if msg.Kind == "cached" {
			cid := string(msg.Data)
			if val, ok := g.pendingUploads[cid]; ok {
				val.lock.Lock()
				*val.res.NumberCached++
				for i, _ := range val.res.CacheNodes {
					if val.res.CacheNodes[i].PeerId == msg.From {
						val.res.CacheNodes[i].Cached = true
					}
				}
				val.lock.Unlock()
				val.Notify <- struct{}{}
			}
		}
		if msg.Kind == "pinned" {
			cid := string(msg.Data)
			if val, ok := g.pendingUploads[cid]; ok {
				val.lock.Lock()
				*val.res.NumberStored++
				for i, _ := range val.res.StorageNodes {
					if val.res.StorageNodes[i].PeerId == msg.From {
						val.res.StorageNodes[i].Stored = true
					}
				}
				val.lock.Unlock()
				val.Notify <- struct{}{}

			}
		}
	}
}

/*
 * Returns after timeout or after at least one node
 * we trust has confirmed the pin
 */
func (g *Gateway) onceUploadRoute(c *gin.Context) {

	timeout_duration := 30 * time.Second

	if g.checkAccessToken(c) {
		return
	}

	net := g.getNetwork()
	if *net.NumberStores == 0 {
		c.String(500, "Not enough Storge nodes configured!")
		return
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
		return
	}

	check.res.Cid = cid
	g.pendingUploads[cid] = check
	ticker := time.NewTicker(timeout_duration)

	// wait until we got feedback from others
	select {
	case <-ticker.C:
		// got timeout
		ticker.Stop()
		check.lock.Lock()
		g.log.WithField("cid", check.res.Cid).Warn("Once has reached timeout")
		check.res.Status = "Timeout"
		check.lock.Unlock()
		break

	case <-check.Notify:
		check.lock.Lock()
		if *check.res.NumberStored >= 1 {
			check.res.Status = "Success"
			check.lock.Unlock()
			break
		}
		check.lock.Unlock()

	}

	// if we are here either success or timeout
	// either way, ship response

	check.lock.Lock()
	defer check.lock.Unlock()
	c.JSON(200, check.res)
}

func (g *Gateway) checkAccessToken(c *gin.Context) bool {
	if len(g.accessTokens) != 0 {
		tokenH := c.Request.Header["Token"]
		if len(tokenH) == 0 {
			c.String(401, "need upload token")
			return true
		}
		token := tokenH[0]
		pass := false
		for _, t := range g.accessTokens {
			if t.Token == token {
				pass = true
			}
		}
		if pass == false {
			c.String(401, "Invalid token")
			return true
		}
	}
	return false
}

func (g *Gateway) checkUploadToken(c *gin.Context) bool {
	if len(g.accessTokens) != 0 {
		tokenH := c.Request.Header["Token"]
		if len(tokenH) == 0 {
			c.String(401, "need upload token")
			return true
		}
		token := tokenH[0]
		pass := false
		for _, t := range g.accessTokens {
			if t.Token == token {
				pass = true
			}
		}
		if pass == false {
			c.String(401, "Invalid token")
			return true
		}
	}
	return false
}

func (g *Gateway) storeFile(c *gin.Context) (string, bool) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(500, err.Error())
		return "", true
	}

	f, err := file.Open()
	if err != nil {
		c.String(500, err.Error())
		return "", true
	}
	cid, err := g.net.UploadAndPin(f)
	if err != nil {
		c.String(500, err.Error())
		return "", true
	}
	return cid, false
}

func (g *Gateway) broadcastCache(cid string) {
	msg := network.PubSubMessage{
		Kind: "cached",
		Data: []byte(cid),
	}
	g.net.SendMessage(&msg)
}

func getType(buf []byte) string {
	// TODO
	return ""
}

type UploadResponse struct {
	Cid          string `json:",omitempty"`
	CacheNodes   []CacheNode
	StorageNodes []StorageNode
	NumberCaches *int   `json:",omitempty"`
	NumberStores *int   `json:",omitempty"`
	NumberCached *int   `json:",omitempty"`
	NumberStored *int   `json:",omitempty"`
	Status       string `json:",omitempty"`
}

type StorageNode struct {
	Name         string `json:",omitempty"`
	Organization string `json:",omitempty"`
	Contact      string `json:",omitempty"`
	Comment      string `json:",omitempty"`
	PeerId       string `json:",omitempty"`
	Stored       bool   `json:",omitempty"`
}

type CacheNode struct {
	Name         string `json:",omitempty"`
	Organization string `json:",omitempty"`
	Contact      string `json:",omitempty"`
	Comment      string `json:",omitempty"`
	PeerId       string `json:",omitempty"`
	Cached       bool   `json:",omitempty"`
}

type PendingUpload struct {
	Notify chan struct{}
	lock   *sync.Mutex
	res    *UploadResponse
}

func intptr(i int) *int {
	a := i
	return &a
}
