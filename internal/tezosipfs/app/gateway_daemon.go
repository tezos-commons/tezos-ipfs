package app

import (
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/network"
)

func (g *Gateway) broadcastCache(cid string) {
	msg := network.PubSubMessage{
		Kind: "cached",
		Data: []byte(cid),
	}
	g.net.SendMessage(&msg)
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
