package p2p

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

// p2p 层 stream 传输回归测试：hdr + stream 头 + 数据在一条连接上按序传输。
// 复现/防回归 gob bufio 预读问题——解码与 stream 数据必须共用同一个
// bufio.Reader（生产路径中由 TCPPeer.br + handleConn 保证）。
func TestSmoke_StreamTransfer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()

	got := make(chan []byte, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		defer conn.Close()
		// 与 TCPPeer.br 相同的模式：解码与 stream 数据共用持久 bufio.Reader
		br := bufio.NewReader(conn)
		dec := &DefaultDecoder{}
		var pending *MessageStoreFile
		for {
			var rpc RPC
			if err := dec.Decode(br, &rpc); err != nil {
				t.Errorf("B decode: %v", err)
				return
			}
			if rpc.Stream {
				if pending == nil {
					t.Errorf("B: no pending header")
					return
				}
				data := make([]byte, pending.Size)
				if _, err := io.ReadFull(io.LimitReader(br, pending.Size), data); err != nil {
					t.Errorf("B read stream: %v", err)
					return
				}
				got <- data
				return
			}
			switch hdr := rpc.Payload.(type) {
			case MessageStoreFile:
				pending = &hdr
			default:
				t.Errorf("B unexpected payload %T", rpc.Payload)
				return
			}
		}
	}()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	enc := &DefaultEncoder{}
	payload := []byte("0123456789abcdef")
	if err := enc.Encode(conn, &RPC{Payload: MessageStoreFile{Key: "k", Size: int64(len(payload))}}); err != nil {
		t.Fatalf("hdr: %v", err)
	}
	if err := enc.Encode(conn, &RPC{Stream: true, StreamSize: int64(len(payload))}); err != nil {
		t.Fatalf("stream: %v", err)
	}
	if _, err := conn.Write(payload); err != nil {
		t.Fatalf("data: %v", err)
	}

	select {
	case d := <-got:
		if !bytes.Equal(d, payload) {
			t.Fatalf("data mismatch: %q", d)
		}
		t.Logf("stream transfer OK, %d bytes", len(d))
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting stream data")
	}
}
