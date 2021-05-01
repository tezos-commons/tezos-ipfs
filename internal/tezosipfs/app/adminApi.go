package app

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/cache"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/db"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/network"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/swarm"
	"strconv"
)

type Admin struct {
	swarm *swarm.Swarm
	db    *db.StormDB
	net   network.NetworkInterface
	log   *logrus.Entry
	c *config.Config
	pin *PinManager
	gateway *Gateway
	cache *cache.S3Cache
}

func NewAdminAPI(s *swarm.Swarm, db *db.StormDB, net network.NetworkInterface, l *logrus.Entry, c *config.Config,pin *PinManager, gateway *Gateway, cache *cache.S3Cache) *Admin {
	a := Admin{
		swarm: s,
		db: db,
		net: net,
		log: l.WithField("source","admin-api"),
		c: c,
		pin: pin,
		gateway: gateway,
		cache: cache,
	}
	return &a
}


func (a *Admin) Run() {
	a.log.Info("Starting admin api on: " + a.c.Admin.Host + ":" + strconv.Itoa(a.c.Admin.Port))
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/pin/:cid",a.pinRequest)
	r.DELETE("/pin/:cid",a.unPinReuest)
	r.POST("/pin/:cid/block",a.blockRequest)
	r.GET("/id",a.idRequest)
	r.Run(a.c.Admin.Host + ":" + strconv.Itoa(a.c.Admin.Port))
}

func (a *Admin) pinRequest(c *gin.Context){
	cid := c.Param("cid")
	if a.pin != nil {
		a.pin.Pin(cid)
	}
	c.String(200, "ok")
}

func (a *Admin) unPinReuest(c *gin.Context){
	cid := c.Param("cid")
	if a.pin != nil {
		a.pin.UnPin(cid)
	}
	if a.cache != nil {
		a.cache.Uncache(cid)
	}
	c.String(200, "ok")
}

func (a *Admin) blockRequest(c *gin.Context){
	cid := c.Param("cid")
	if a.pin != nil {
		a.pin.Block(cid)
	}
	if a.cache != nil {
		a.cache.Uncache(cid)
	}
	c.String(200, "ok")
}


func (a *Admin) idRequest(c *gin.Context){
	c.String(200,a.net.ID())
}