package swarm

import (
	"encoding/json"
	"fmt"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/network"
)

type PeerAdvertisment struct {
	Name string
	Organization string
	Contact string
	Comment string
	PeerId string
	TrustedPeers []string
	CacheFor []string
	PinFor []string
}

func (p *PeerAdvertisment) toTransportFormat() *network.PubSubMessage {
	b,e := json.Marshal(*p)
	if e != nil {
		fmt.Println(e)
	}
	res := network.PubSubMessage{
		Data: b,
		Kind: "peer_advertisement",
	}
	return &res
}


type PinRequest struct {
	Cid string
	Data []byte // if small, files, distribute via pubsub directly
                // a small file is <= 256kb, so most json files, metadata etc
}