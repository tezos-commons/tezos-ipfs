package app

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"sync"
)

func (g *Gateway) cacheFile(cid string) {
	reader, err := g.net.GetFile(context.Background(), cid)
	if err != nil {
		g.log.Error(err)
	}
	// todo split reader if possible
	buf, _ := ioutil.ReadAll(reader)
	if g.cache != nil {
		g.cache.StoreFile(cid, bytes.NewReader(buf))
		g.log.WithField("cid", cid).Trace("Stored in cache")
		g.broadcastCache(cid)
	} else {
		g.log.WithField("cid", cid).Error("got cache request, but have no cache configured...")
	}
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

type UploadResponse struct {
	Cid          string        `json:",omitempty"`
	CacheNodes   []CacheNode   `json:",omitempty"`
	StorageNodes []StorageNode `json:",omitempty"`
	NumberCaches *int          `json:",omitempty"`
	NumberStores *int          `json:",omitempty"`
	NumberCached *int          `json:",omitempty"`
	NumberStored *int          `json:",omitempty"`
	Status       string        `json:",omitempty"`
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
	ID     string
	Notify chan struct{}
	lock   *sync.Mutex
	res    *UploadResponse
}

func getType(buf []byte) string {
	// TODO
	return ""
}

func intptr(i int) *int {
	a := i
	return &a
}
