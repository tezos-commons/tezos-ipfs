package db

import (
	"github.com/asdine/storm/v3"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/common"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
)

type StormDB struct {
	log   *logrus.Entry
	storm *storm.DB
}

func NewStormDB(c *config.Config, l *logrus.Entry) *StormDB {
	d := StormDB{}
	d.log = l.WithField("source", "boltdb")

	db, err := storm.Open(c.DB.Storm)
	if err != nil {
		d.log.Fatal(err)
		return nil
	}

	d.storm = db

	return &d
}

func (d *StormDB) Write(bucketName, key, value []byte) {
	skey := append(bucketName, key...)
	obj := common.KeyValue{
		ID:    10,
		Key:   skey,
		Value: value,
	}

	d.storm.Save(&obj)
}

func (d *StormDB) Get(bucketName, key []byte) (val []byte, length int) {
	skey := append(bucketName, key...)
	obj := common.KeyValue{}
	d.storm.One("Key", skey, &obj)
	return obj.Value, len(obj.Value)
}

func (d *StormDB) SavePin(p *common.Pin) error {
	return d.storm.Save(p)
}

func (d *StormDB) RemovePin(p *common.Pin) error {
	return d.storm.DeleteStruct(p)
}

func (d *StormDB) GetPin(cid string) (*common.Pin, error) {
	obj := common.Pin{}
	e := d.storm.One("Cid", cid, &obj)
	return &obj, e
}

func (d *StormDB) PaginatedGetAllPin(pagesize, page int) ([]common.Pin, error) {
	var pins []common.Pin
	err := d.storm.Range("ID", pagesize*(page-1), pagesize*page, &pins, storm.Reverse())
	return pins, err
}

func (d *StormDB) IsBlocked(cid string) bool {
	p, err := d.GetPin(cid)
	if err != nil {
		return false
	}
	if p.Status == "blocked" {
		return true
	}
	return false
}
