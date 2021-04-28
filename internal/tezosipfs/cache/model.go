package cache

import "io"

type Cache interface {
	GetFile(cid string) (int64,io.Reader,error)
	StoreFile(cid string, reader io.ReadSeeker)
}
