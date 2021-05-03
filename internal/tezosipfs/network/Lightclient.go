package network

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	dht2 "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
	"time"
)

type Lightclient struct {
	client           *ipfslite.Peer
	log              *logrus.Entry
	privkey          []byte
	h                host.Host
	connected        map[string]bool
	dht              *dht2.IpfsDHT
	pubsub           *pubsub.PubSub
	floodsub         *pubsub.PubSub
	topic            *pubsub.Topic
	ftopic           *pubsub.Topic
	pubsubscriptions []chan *PubSubMessage
	msgcache         *lru.Cache
	cm               *connmgr.BasicConnMgr
}

var options = []libp2p.Option{
	libp2p.NATPortMap(),
	libp2p.EnableAutoRelay(),
	libp2p.EnableNATService(),
	libp2p.Transport(libp2pquic.NewTransport),
	libp2p.EnableAutoRelay(),
	libp2p.DefaultTransports,
}

func NewLightclient(privkey []byte, log *logrus.Entry) *Lightclient {
	l := Lightclient{}
	l.privkey = privkey
	l.log = log.WithField("source", "light_client")
	return &l
}

func (l *Lightclient) Setup() {
	cm := connmgr.NewConnManager(20, 50, time.Minute)
	options = append(options, libp2p.ConnectionManager(cm))
	ctx, _ := context.WithCancel(context.Background())
	ds, err := ipfslite.BadgerDatastore("/tmp/badger")
	if err != nil {
		l.log.Fatal(err)
	}
	priv, err := crypto.UnmarshalPrivateKey(l.privkey)
	if err != nil {
		panic(err)
	}
	listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005/")
	listen2, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/udp/4005/")
	l.connected = map[string]bool{}
	h, _, err := ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{listen, listen2},
		ds,
		options...,
	)
	ps, err := pubsub.NewGossipSub(ctx, h)
	ps2, err := pubsub.NewFloodSub(ctx, h)
	if err != nil {
		panic(err)
	}
	l.pubsub = ps
	l.floodsub = ps2
	topic, err := ps.Join(BROADCAST_TOPIC)
	topic2, err := ps2.Join(BROADCAST_TOPIC)
	if err != nil {
		panic(err)
	}
	l.topic = topic
	l.ftopic = topic2
	dht := dht2.NewDHT(context.Background(), h, ds)
	l.dht = dht
	if err != nil {
		panic(err)
	}
	lite, err := ipfslite.New(ctx, ds, h, dht, nil)
	if err != nil {
		l.log.Fatal(err)
	}
	l.msgcache, _ = lru.New(1500)
	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())
	l.client = lite
	l.h = h
	l.cm = cm

	l.log.Info("My peerID is: ", h.ID().String())
}

func (l *Lightclient) GetFile(ctxorig context.Context, cidStr string) (io.Reader, error) {
	ctx, _ := context.WithCancel(ctxorig)
	l.log.Trace("Get File: " + cidStr)
	c, err := cid.Decode(cidStr)
	if err != nil {
		l.log.Error(err)
		return nil, err
	}
	rsc, err := l.client.GetFile(ctx, c)
	if err != nil {
		l.log.Error(err)
		return nil, err
	}
	return rsc, nil
}

func (l *Lightclient) AddFile(source io.Reader) (string, error) {
	node, err := l.client.AddFile(context.Background(), source, nil)
	if err != nil {
		l.log.Error("Error adding file ", err)
		return "", err
	}
	return node.Cid().String(), nil
}

func (l *Lightclient) Connect(peers []string) error {
	connected := make(chan struct{})
	ctx := context.Background()
	var wg sync.WaitGroup
	for _, peerString := range peers {
		if peerString == l.h.ID().String() {
			continue
		}
		p, err := peer.Decode(peerString)
		if err != nil {
			l.log.Warn("can not parse peerID: ", err)
			continue
		}
		if l.cm.IsProtected(p, "tezos-ipfs") {
			continue
		}
		go l.cm.Protect(p, "tezos-ipfs")
		pinfo, err := l.dht.FindPeer(ctx, p)
		if err != nil {
			l.log.Warn("error creating pinfo: ", err)
			continue
		}

		l.h.Peerstore().AddAddrs(pinfo.ID, pinfo.Addrs, peerstore.PermanentAddrTTL)
		wg.Add(1)
		go func(pinfo peer.AddrInfo) {
			defer wg.Done()
			err := l.h.Connect(ctx, pinfo)
			if err != nil {
				l.log.Warn(err)
			} else {
				if _, ok := l.connected[pinfo.ID.String()]; !ok {
					l.log.Info("Connected with ", pinfo.ID)
					l.connected[pinfo.ID.String()] = true
				}
			}
			connected <- struct{}{}
		}(pinfo)
	}

	go func() {
		wg.Wait()
		close(connected)
	}()
	go l.listenPubsub()

	return nil
}

func (l *Lightclient) SendMessage(msg *PubSubMessage) {
	msg.Id = uuid.New().String()
	data, _ := json.Marshal(msg)
	l.topic.Publish(context.Background(), data)
	l.ftopic.Publish(context.Background(), data)
}

func (l *Lightclient) listenPubsub() {
	s, e := l.topic.Subscribe()
	if e != nil {
		l.log.Fatal(e)
	}
	for {
		msg, e := s.Next(context.Background())
		if e != nil {
			l.log.Warn(e)
		}
		p, _ := peer.IDFromBytes(msg.From)
		psmg := PubSubMessage{}
		json.Unmarshal(msg.Data, &psmg)

		if _, ok := l.msgcache.Get(psmg.Id); !ok {
			l.msgcache.Add(psmg.Id, true)
			psmg.From = p.String()
			for _, c := range l.pubsubscriptions {
				c <- &psmg
			}
		}
	}
}

func (l *Lightclient) Subscribe() chan *PubSubMessage {
	res := make(chan *PubSubMessage, 10)
	l.pubsubscriptions = append(l.pubsubscriptions, res)
	return res
}

func (l *Lightclient) ID() string {
	return l.h.ID().String()
}

func (l *Lightclient) UploadAndPin(file io.Reader) (string, error) {
	fnode, err := l.client.AddFile(context.Background(), file, nil)
	if err != nil {
		return "", err
	}
	c, _ := cid.Decode(fnode.Cid().String())
	pinRequest := PubSubMessage{
		Data: []byte(c.Hash().B58String()),
		Kind: "new_object",
	}
	l.SendMessage(&pinRequest)
	l.log.WithField("cid", fnode.Cid()).Trace("sending pin request")
	return c.Hash().B58String(), err
}

func (l *Lightclient) LocalPin(cid string) error {
	l.log.Error("attempted to pin something on lightclient....")
	return nil
}

func (l *Lightclient) RemovePin(cid string) error {
	l.log.Error("attempted to unpin something on lightclient....")
	return nil
}
