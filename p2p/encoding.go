package p2p

import (
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type Encoder interface {
	Encode(io.Writer, *RPC) error
}

type DefaultDecoder struct {
}

type DefaultEncoder struct {
}

func (e DefaultEncoder) Encode(w io.Writer,rpc *RPC) error {
	_,err:=w.Write(rpc.Payload)
	if err!=nil{
		return err
	}
	return nil
}

func (d DefaultDecoder) Decode(r io.Reader, rpc *RPC) error {
	buffer := make([]byte, 1024)
	n, err := r.Read(buffer)
	if err != nil {
		return err
	}
	copbuf := make([]byte, n)
	copy(copbuf, buffer[:n])
	rpc.Payload = copbuf
	return nil
}
