package network

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"io"
)

type IPFS struct {
	sh *shell.Shell
	log *logrus.Entry
	connected map[string]bool
	pubsubscriptions []chan *PubSubMessage
	id string
	msgcache *lru.Cache
}


func NewIPFS(c *config.Config,l *logrus.Entry) *IPFS{
	url := c.GetIpfsAPI()
	if url == nil {
		return nil
	}
	r := IPFS{}
	r.connected = map[string]bool{}
	r.pubsubscriptions = []chan *PubSubMessage{}
	r.log = l.WithField("source","ipfs-wrapper")
	r.log.Info("Connecting to external IPFS Node....")
	sh := shell.NewShell(*url)
	if sh == nil {
		r.log.Fatal("Can not connect to IPFS via " + *url)
	}
	r.sh = sh
	r.msgcache,_ = lru.New(1500)
	pi,_ := r.sh.ID()
	r.id = pi.ID
	go r.listenPubSub()
	return &r
}

func (i *IPFS) Listen() chan *shell.Message {
	s,e := i.sh.PubSubSubscribe(BROADCAST_TOPIC)
	if e != nil {
		i.log.Fatal(e)
		return nil
	}

	res := make(chan *shell.Message)
	go func() {
		for {
			msg,e := s.Next()
			if e != nil {
				i.log.Warn("Error getting PubSub: " + e.Error())
			} else {
				res <- msg
			}
		}
	}()

	return res
}

func (i *IPFS) GetFile(ctx context.Context, cidStr string) (io.Reader,error) {
	resp, err := i.sh.Request("get", cidStr).Option("create", true).Send(context.Background())
	if err != nil {
		i.log.Error("Cant get file ", err)
	}
	return resp.Output,nil
}

func (i *IPFS) Connect(peers []string) error {
	ctx := context.Background()
	for _,a := range peers {
		pi,err := i.sh.FindPeer(a)
		if err != nil {
			i.log.Trace("can not parse peerID: ", err)
			continue
		}
		for _,pa := range pi.Addrs {
			err = i.sh.SwarmConnect(ctx, pa + "/p2p/" + pi.ID)
			if err != nil {
				i.log.Trace("can not connect to: ", err)
				continue
			} else {
				if _,ok := i.connected[pi.ID]; !ok {
					i.log.Info("Connected with ", pi.ID)
					i.connected[pi.ID] = true
				}
			}
		}
	}
	return nil
}


func (i *IPFS) SendMessage(msg *PubSubMessage) {
	msg.Id = uuid.New().String()
	data,_ := json.Marshal(msg)
	i.sh.PubSubPublish(BROADCAST_TOPIC,string(data))
}


func (i *IPFS) listenPubSub(){
	s,e := i.sh.PubSubSubscribe(BROADCAST_TOPIC)
	if e != nil {
		i.log.Fatal(e)
	}
	for {
		var msg, e = s.Next()
		if e != nil {
			i.log.Warn(e)
		}
		psmg := PubSubMessage{}
		json.Unmarshal(msg.Data,&psmg)
		if _,ok := i.msgcache.Get(psmg.Id); !ok {
			i.msgcache.Add(psmg.Id,true)
			psmg.From = msg.From.String()
			if psmg.From != i.id {
				for _,c := range i.pubsubscriptions {
					c <- &psmg
				}
			}
		}
	}
}

func (i *IPFS) Subscribe() chan *PubSubMessage {
	res := make(chan *PubSubMessage,10)
	i.pubsubscriptions = append(i.pubsubscriptions, res)
	return res
}

func (i *IPFS) UploadAndPin(file io.Reader) (string,error){
	cid,err := i.sh.Add(file)
	if err != nil {
		return cid,err
	}
	err = i.sh.Pin(cid)
	pinRequest := PubSubMessage{
		Data: []byte(cid),
		Kind: "pin_request",
	}
	i.SendMessage(&pinRequest)
	return cid,err
}