package network

import (
	"context"
	"io"
)

type NetworkInterface interface {
	 GetFile(ctx context.Context, cidStr string) (io.Reader,error)
	 Connect(peers []string) error
	 SendMessage(msg *PubSubMessage)
	 Subscribe() chan *PubSubMessage
}

type PubSubMessage struct {
	Id string
	Data []byte
	Kind string
	From string
}