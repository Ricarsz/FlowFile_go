package p2p

import (
	"net"
)

type Peer interface {
	Conn() net.Conn
	Close() error
}

type RPC struct{
	From net.Addr
	Payload []byte
}

type Transport interface {
	ListenAndAccept() error
	Dial(add string) error
	Close() error
	//监听地址
	Addr() string
	Cousume() <-chan RPC
}
