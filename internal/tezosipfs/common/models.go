package common

import "time"

type Pin struct {
	ID      int			`storm:"id,increment"`
	Created time.Time	`storm:"index"`
	Cid     string		`storm:"unique"`
	From    string		`storm:"index"`
	Status  string		`storm:"index"`
	Size    int64
}


type KeyValue struct {
	ID int `storm:"id,increment"`
	Key []byte `storm:"unique"`
	Value []byte
}