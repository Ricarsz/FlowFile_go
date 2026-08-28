package p2p

import (
	"net"
)

type Peer interface {
	Conn() net.Conn
	Close() error
	Send(b []byte) error
}

type RPC struct {
	From    net.Addr
	Payload []byte
}

type Transport interface {
	ListenAndAccept() error
	Dial(add string) error
	Close() error
	Addr() string
	Consume() <-chan RPC
}
