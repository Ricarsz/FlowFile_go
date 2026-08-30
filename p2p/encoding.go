package p2p

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
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

func GOBNewDecoder(r io.Reader, rpc *RPC) error {
	return gob.NewDecoder(r).Decode(&rpc.Payload)
}

func GOBNewEncoder(w io.Writer, rpc *RPC) error {
	return gob.NewEncoder(w).Encode(&rpc.Payload)
}

func (e DefaultEncoder) Encode(w io.Writer, rpc *RPC) error {
	if rpc.Stream {
		if _, err := w.Write([]byte{IncomingStream}); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, rpc.StreamSize); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte{IncomingMessage}); err != nil {
			return err
		}
		if err := GOBNewEncoder(w, rpc); err != nil {
			return err
		}
	}
	return nil
}

func (d DefaultDecoder) Decode(r io.Reader, rpc *RPC) error {
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	messageType := buf[0]
	if messageType == IncomingMessage {
		rpc.Stream = false
		if err := GOBNewDecoder(r, rpc); err != nil {
			return err
		}
	} else if messageType == IncomingStream {
		rpc.Stream = true
		if err := binary.Read(r, binary.LittleEndian, &rpc.StreamSize); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown message type %x", messageType)
	}
	return nil
}
