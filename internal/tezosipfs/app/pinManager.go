package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/network"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/swarm"
	"io"
	"io/ioutil"
	"strconv"
	"time"
)

type PinManager struct {
	swarm *swarm.Swarm
	db *bolt.DB
	net network.NetworkInterface
	log *logrus.Entry
}


func NewPinManager(s *swarm.Swarm, db *bolt.DB, net network.NetworkInterface, l *logrus.Entry, c *config.Config) *PinManager {
	if c.PinManagerEnabled == false {
		l.Info("PinManager disabled")
		return nil
	}
	pin := PinManager{
		swarm: s,
		db:    db,
		net: net,
		log: l.WithField("source","pin-manager"),
	}

	go pin.listen()
	return &pin
}


func (pin *PinManager) listen(){
	ch := pin.net.Subscribe()
	for {
		msg := <- ch
		if msg.Kind == "new_object" && pin.swarm.PinFor(msg.From) {
			pin.log.WithField("source",msg.From).WithField("cid",string(msg.Data)).Trace("pin request")
			if pin.swarm.PinFor(msg.From) {
				pin.log.WithField("cid",string(msg.Data)).WithField("origin",msg.From).Info("Auto-Pin")
				go pin.Pin(string(msg.Data))
			}
		}
	}
}

func (pin *PinManager) Pin(cid string){
	start := time.Now()
 	existing,err := pin.getPin(cid)
 	if existing != nil || err == nil {
 		pin.log.WithField("cid",cid).Info("already pinned content")
 		pin.broadcastPin(cid)
 		return
	}
	p := &Pin{
		Cid: cid,
		Created: time.Now(),
		Status: "pinning",
	}
	pin.savePin(p)
	err = pin.net.LocalPin(cid)
	if err != nil {
		pin.log.WithField("cid",cid).Error(err)
		p.Status = "Error"
		pin.savePin(p)
		return
	}
	// make sure we have item stored
	tries := 0
	for {
		if tries >= 20 {
			break
		}
		pin.log.WithField("cid",cid).Trace("Pin try: ", strconv.Itoa(tries+1))
		f,e := pin.net.GetFile(context.Background(),cid)
		if e == nil {
			// just make sure we have entire file
			count,_ := io.Copy(ioutil.Discard,f)
			pin.log.WithField("cid",cid).WithField("size",count).WithField("duration",time.Since(start)).Info("Store completed")
			p.Status = "pinned"
			pin.savePin(p)
			pin.broadcastPin(cid)
			break
		} else {
			pin.log.WithField("cid",cid).Warn(err)
		}
		time.Sleep(3*time.Second)
		tries++
	}
	if tries == 20 {
		p.Status = "timout"
		pin.savePin(p)
		pin.log.WithField("cid",cid).Warn("Could not complete pin, will try again later")
	}
}

func (pin *PinManager) broadcastPin(cid string) {
	msg := network.PubSubMessage{
		Kind: "pinned",
		Data: []byte(cid),
	}
	pin.net.SendMessage(&msg)
}

func (pin *PinManager) savePin(p *Pin) error {
	return pin.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Pins"))
		buf, err := json.Marshal(pin)
		if err != nil {
			return err
		}
		return b.Put([]byte(p.Cid), buf)
	})
}

func (pin *PinManager) getPin(cid string) (p *Pin,err error) {
	err = pin.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Pins"))
		bytes := b.Get([]byte(cid))
		if len(bytes) == 0 {
			return errors.New("not found")
		}
		json.Unmarshal(bytes,p)
		return nil
	})
	return p,err
}






type Pin struct {
	Created time.Time
	Cid string // key
	From string
	Status string
}
