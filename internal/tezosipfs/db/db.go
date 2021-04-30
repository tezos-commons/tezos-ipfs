package db

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"time"
)

type BoltDb struct {
	log *logrus.Entry
	bolt *bolt.DB
}

func NewBoltDb(c *config.Config,l *logrus.Entry) (*BoltDb, *bolt.DB) {
	d := BoltDb{}
	d.log = l.WithField("source","boltdb")

	db, err := bolt.Open(c.DB.Bolt, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil,nil
	}

	tx, err := db.Begin(true)
	if err != nil {
		d.log.Error(err)
	}
	defer tx.Rollback()
	tx.CreateBucket([]byte("Config"))
	tx.CreateBucket([]byte("Peers"))
	tx.CreateBucket([]byte("Pins"))
	tx.Commit()

	d.bolt = db

	return &d,db
}

func (d *BoltDb) Write(bucketName, key, value []byte) {
	err := d.bolt.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		err = bkt.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		d.log.Fatal(err)
	}
}


func (d *BoltDb) Get(bucketName, key []byte) (val []byte, length int) {
	err := d.bolt.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(bucketName)
		if bkt == nil {
			return fmt.Errorf("Bucket %q not found!", bucketName)
		}
		val = bkt.Get(key)
		return nil
	})
	if err != nil {
		d.log.Fatal(err)
	}
	return val, len(string(val))
}