package p2p

import "encoding/gob"

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

func init() {
	gob.Register(MessagePeerList{})
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
	gob.Register(Message{})
}

type MessagePeerList struct {
	Addrs []string
}

type MessageStoreFile struct {
	Key  string
	Size int64
}
type MessageGetFile struct {
	Key string
}

type Message struct {
	Payload interface{}
}
