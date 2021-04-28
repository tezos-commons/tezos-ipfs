package swarm

import (
	"encoding/json"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
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
	addrs      []string // array of peerIds.
	pids       []peer.ID
	knownPeers map[string]*PeerAdvertisment
	l          *sync.Mutex
	log        *logrus.Entry
	net        network.NetworkInterface
	config     *config.Config
}

func NewSwarm(c *config.Config, l *logrus.Entry, net network.NetworkInterface) *Swarm {
	s := Swarm{}
	s.log = l.WithField("source","swarm")
	s.l = &sync.Mutex{}
	s.knownPeers = map[string]*PeerAdvertisment{}
	s.addrs = []string{}
	s.net = net
	s.config = c
	s.pids = []peer.ID{}
	s.log.Trace("init")
	go s.updateConfig(c)
	go s.periodic()
	go s.advertiseMyself()
	go s.subscribe()
	return &s
}

func (s *Swarm) updateConfig(c *config.Config){
	ch := c.GetUpdates()
	for {
		newConfig := <- ch
		newaddrs := []string{}
		newpids := []peer.ID{}
		for _,a := range newConfig.Peers.CacheFor {
			newaddrs = append(s.addrs, a)
		}
		for _,a := range newConfig.Peers.PinFor {
			newaddrs = append(s.addrs, a)
		}
		for _,a := range newConfig.Peers.TrustedPeers {
			newaddrs = append(s.addrs, a)
		}
		for _,a := range unique(newaddrs) {
			ma,e := multiaddr.NewMultiaddr(a)
			if e != nil {
				continue
			}
			ai,e := peer.AddrInfoFromP2pAddr(ma)
			if e != nil {
				continue
			}
			newpids = append(s.pids, ai.ID)
		}

		s.l.Lock()
		s.addrs = unique(newaddrs)
		s.pids = newpids
		s.l.Unlock()
		go s.connect()
	}

}

func (s *Swarm) advertiseMyself() {
	msg := PeerAdvertisment{
		Name: s.config.Identity.Name,
		Organization: s.config.Identity.Organization,
		Contact: s.config.Identity.Contact,
		Comment: s.config.Identity.Comment,
	}
	for {
		time.Sleep(5*time.Second)
		msg.TrustedPeers = s.config.Peers.TrustedPeers
		msg.CacheFor = s.config.Peers.CacheFor
		msg.PinFor = s.config.Peers.PinFor
		s.net.SendMessage(msg.toTransportFormat())
	}
}


func (s *Swarm) connect(){
	s.l.Lock()
	defer s.l.Unlock()
	s.net.Connect(s.addrs)
}

func (s *Swarm) periodic(){
	for {
		time.Sleep(15*time.Second)
		s.net.Connect(s.addrs)
	}
}

func (s *Swarm) subscribe(){
	ch := s.net.Subscribe()
	for {
		msg := <- ch
		if msg.Kind == "peer_advertisement" {
			padv := PeerAdvertisment{}
			json.Unmarshal(msg.Data,&padv)
			padv.PeerId = msg.From
			s.l.Lock()
			s.knownPeers[padv.PeerId] = &padv
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