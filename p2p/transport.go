package p2p

import (
	"io"
	"net"
)

type Peer interface {
	Conn() net.Conn
	Reader() io.Reader
	Close() error
	Send(b []byte) error
	CloseStream() error
}

type RPC struct {
	From       string
	Payload    interface{}
	Stream     bool
	StreamSize int64
}

type Transport interface {
	ListenAndAccept() error
	Dial(add string) error
	Close() error
	Addr() string
	Consume() <-chan RPC
}
