package swarm

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/network"
	"sync"
	"time"
)

/*
 * swarm manages libp2p connections to
 * the nodes we care about
 */
type Swarm struct {
	trustedPeers  []string // directly or 2nd level peers
	knownPeers    map[string]*PeerAdvertisment
	knownPeersTTL map[string]time.Time // stores last update
	cacheFor      []string
	pinFor        []string
	l             *sync.Mutex
	log           *logrus.Entry
	net           network.NetworkInterface
	config        *config.Config
}

func NewSwarm(c *config.Config, l *logrus.Entry, net network.NetworkInterface) *Swarm {
	s := Swarm{}
	s.log = l.WithField("source", "swarm")
	s.l = &sync.Mutex{}
	s.knownPeers = map[string]*PeerAdvertisment{}
	s.knownPeersTTL = map[string]time.Time{}
	s.trustedPeers = []string{}
	s.net = net
	s.config = c
	s.log.Trace("init")
	go s.updateConfig(c)
	go s.periodic()
	go s.advertiseMyself()
	go s.subscribe()
	go s.purgePeers()
	return &s
}

func (s *Swarm) updateConfig(c *config.Config) {
	ch := c.GetUpdates()
	for {
		newConfig := <-ch
		newaddrs := []string{}
		for _, a := range newConfig.Peers.CacheFor {
			newaddrs = append(newaddrs, a)
		}
		for _, a := range newConfig.Peers.PinFor {
			newaddrs = append(newaddrs, a)
		}
		for _, a := range newConfig.Peers.TrustedPeers {
			newaddrs = append(newaddrs, a)
		}

		s.l.Lock()
		s.pinFor = newConfig.Peers.PinFor
		s.cacheFor = newConfig.Peers.CacheFor
		s.trustedPeers = unique(newaddrs)
		s.l.Unlock()
		go s.connect()
	}

}

func (s *Swarm) advertiseMyself() {
	msg := PeerAdvertisment{
		Name:         s.config.Identity.Name,
		Organization: s.config.Identity.Organization,
		Contact:      s.config.Identity.Contact,
		Comment:      s.config.Identity.Comment,
	}
	for {
		time.Sleep(5 * time.Second)
		msg.TrustedPeers = s.config.Peers.TrustedPeers
		msg.CacheFor = s.config.Peers.CacheFor
		msg.PinFor = s.config.Peers.PinFor
		s.net.SendMessage(msg.toTransportFormat())
	}
}

func (s *Swarm) connect() {
	s.l.Lock()
	defer s.l.Unlock()
	cp := s.trustedPeers
	go s.net.Connect(cp)
}

func (s *Swarm) periodic() {
	for {
		time.Sleep(15 * time.Second)
		s.net.Connect(s.trustedPeers)
	}
}

func (s *Swarm) purgePeers() {
	// makes sure dead peers are removed from our list
	timeout := 20 * time.Second
	for {
		time.Sleep(10 * time.Second)
		for id, t := range s.knownPeersTTL {
			if time.Now().Add(-1 * timeout).After(t) {
				s.l.Lock()
				delete(s.knownPeersTTL, id)
				delete(s.knownPeers, id)
				s.l.Unlock()
			}
		}
	}
}

func (s *Swarm) subscribe() {
	ch := s.net.Subscribe()
	for {
		msg := <-ch
		if msg.Kind == "peer_advertisement" {
			padv := PeerAdvertisment{}
			json.Unmarshal(msg.Data, &padv)
			padv.PeerId = msg.From
			s.l.Lock()
			s.knownPeers[padv.PeerId] = &padv
			s.knownPeersTTL[padv.PeerId] = time.Now()
			for _, trusted := range s.config.Peers.TrustedPeers {
				if trusted == padv.PeerId {
					s.trustedPeers = append(s.trustedPeers, padv.TrustedPeers...)
					s.trustedPeers = unique(s.trustedPeers)
				}
			}
			s.l.Unlock()
		}
	}
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (s *Swarm) PinFor(pid string) bool {
	s.l.Lock()
	defer s.l.Unlock()
	for _, a := range s.pinFor {
		if a == pid {
			return true
		}
	}
	return false
}

func (s *Swarm) CacheFor(pid string) bool {
	s.l.Lock()
	defer s.l.Unlock()
	for _, a := range s.cacheFor {
		if a == pid {
			return true
		}
	}
	return false
}

func (s *Swarm) IsTrusted(pid string) bool {
	s.l.Lock()
	defer s.l.Unlock()
	for _, a := range s.trustedPeers {
		if a == pid {
			return true
		}
	}
	for _, a := range s.cacheFor {
		if a == pid {
			return true
		}
	}
	for _, a := range s.pinFor {
		if a == pid {
			return true
		}
	}
	return false
}

func (s *Swarm) CacheForUs() []*PeerAdvertisment {
	res := []*PeerAdvertisment{}
	id := s.net.ID()
	for _, a := range s.knownPeers {
		for _, b := range a.CacheFor {
			if b == id {
				res = append(res, a)
			}
		}
	}
	return res
}

func (s *Swarm) PinForUs() []*PeerAdvertisment {
	res := []*PeerAdvertisment{}
	id := s.net.ID()
	for _, a := range s.knownPeers {
		for _, b := range a.PinFor {
			if b == id {
				res = append(res, a)
			}
		}
	}
	return res
}
