package app

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/cache"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/network"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/swarm"
	"io/ioutil"
	"strconv"
	"sync"
)

type Gateway struct {
	log   *logrus.Entry
	net   network.NetworkInterface
	cache cache.Cache
	port  int
	swarm *swarm.Swarm
	accessTokens []config.AccessTokens
	uploadTokens []config.AccessTokens
	l *sync.Mutex
}

func NewGateway(c *config.Config,net network.NetworkInterface,l *logrus.Entry,s *swarm.Swarm) *Gateway {
	if !c.GatewayEnabled {
		l.Info("HTTP Gateway disabled")
		return nil
	}
	g := Gateway{}
	g.net = net
	g.swarm = s
	g.l = &sync.Mutex{}
	g.log = l.WithField("source","gateway")
	g.port = c.Gateway.Server.Port
	if c.Gateway.Storage.S3.Bucket != "" {
		g.log.Info("Using S3 as storage backend")
		cache := cache.NewS3Cache(c,l)
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
	return &g
}


func (g *Gateway) watchConfig(c *config.Config){
	ch := c.GetUpdates()
	for {
		n := <- ch
		g.l.Lock()
		g.log.Info("updating config")
		g.accessTokens = n.Gateway.Server.AccessTokens
		g.uploadTokens = n.Gateway.Server.UploadToken
		g.l.Unlock()
	}
}


func (g *Gateway) Run(){
	g.log.Info("Starting gateway on port :" + strconv.Itoa(g.port))
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/ipfs/:cid", g.ipfsRoute)
	r.POST("/upload", g.uploadRoute)
	r.Run("0.0.0.0:" + strconv.Itoa(g.port))
}

func (g *Gateway) ipfsRoute (c *gin.Context) {

	if len(g.accessTokens) != 0 {
		tokenH := c.Request.Header["Token"]
		if len(tokenH) == 0 {
			c.String(401,"need upload token")
			return
		}
		token := tokenH[0]
		pass :=  false
		for _,t := range g.accessTokens {
			if t.Token == token {
				pass = true
			}
		}
		if pass == false {
			c.String(401,"Invalid token")
			return
		}
	}

	cid := c.Param("cid")
	headers := map[string]string{}
	if len(cid) <= 12 || len(cid) >= 64 {
		c.String(500,"invalid cid")
		return
	}
	l,reader,err := g.cache.GetFile(cid)
	if err == nil {
		g.log.WithField("cid",cid).Trace("Cache hit")
		buf,_ := ioutil.ReadAll(reader)
		c.DataFromReader(200,l, getType(buf),bytes.NewReader(buf),headers)
		return
	}
	reader, err = g.net.GetFile(c,cid)
	if err != nil {
		c.String(404,":(")
		return
	}
	g.log.WithField("cid",cid).Trace("Found via Network")
	buf,_ := ioutil.ReadAll(reader)
	g.cache.StoreFile(cid,bytes.NewReader(buf))
	c.DataFromReader(200,int64(len(buf)), getType(buf),bytes.NewReader(buf),headers)
}



func (g *Gateway) uploadRoute (c *gin.Context) {

	if len(g.accessTokens) != 0 {
		tokenH := c.Request.Header["Token"]
		if len(tokenH) == 0 {
			c.String(401,"need upload token")
			return
		}
		token := tokenH[0]
		pass :=  false
		for _,t := range g.accessTokens {
			if t.Token == token {
				pass = true
			}
		}
		if pass == false {
			c.String(401,"Invalid token")
			return
		}
	}

	file,err := c.FormFile("file")
	if err != nil {
		c.String(500,err.Error())
		return
	}

	f,err := file.Open()
	if err != nil {
		c.String(500,err.Error())
		return
	}
	cid,err := g.net.UploadAndPin(f)
	if err != nil {
		c.String(500,err.Error())
		return
	}

	c.String(200,cid)
}


func getType(buf []byte) string {
	// TODO https://stackoverflow.com/questions/23714383/what-are-all-the-possible-values-for-http-content-type-header
	// kind, _ := filetype.Match(buf)
	return ""
}