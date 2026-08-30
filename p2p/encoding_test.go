package p2p

import (
	"bytes"
	"testing"
)

func TestEncodeDecode_Message(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := DefaultEncoder{}
	dec := DefaultDecoder{}

	// 用最简 Payload 验证不粘包：连续写两条消息，读出来不应错位
	rpc1 := &RPC{Payload: []byte("hello")}
	rpc2 := &RPC{Payload: []byte("world")}

	if err := enc.Encode(buf, rpc1); err != nil {
		t.Fatalf("encode1: %v", err)
	}
	if err := enc.Encode(buf, rpc2); err != nil {
		t.Fatalf("encode2: %v", err)
	}

	// 此时 buf 里是 [0x1+gob(hello)][0x1+gob(world)] 粘在一起
	var out1 RPC
	if err := dec.Decode(buf, &out1); err != nil {
		t.Fatalf("decode1: %v", err)
	}
	if out1.Stream {
		t.Fatalf("out1 should not be stream")
	}
	if string(out1.Payload) != "hello" {
		t.Fatalf("out1 payload = %q, want hello", out1.Payload)
	}

	var out2 RPC
	if err := dec.Decode(buf, &out2); err != nil {
		t.Fatalf("decode2: %v", err)
	}
	if string(out2.Payload) != "world" {
		t.Fatalf("out2 payload = %q, want world", out2.Payload)
	}
}

func TestEncodeDecode_Stream(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := DefaultEncoder{}
	dec := DefaultDecoder{}

	// Stream 头：只编码 type+size，不编码 payload
	streamRPC := &RPC{Stream: true, StreamSize: 1024}
	if err := enc.Encode(buf, streamRPC); err != nil {
		t.Fatalf("encode stream: %v", err)
	}
	// 紧接着一条普通消息，测试 header+size 后再接 message 不错位
	msgRPC := &RPC{Payload: []byte("after stream")}
	if err := enc.Encode(buf, msgRPC); err != nil {
		t.Fatalf("encode msg: %v", err)
	}

	var outStream RPC
	if err := dec.Decode(buf, &outStream); err != nil {
		t.Fatalf("decode stream: %v", err)
	}
	if !outStream.Stream || outStream.StreamSize != 1024 {
		t.Fatalf("outStream = %+v, want Stream true size 1024", outStream)
	}

	var outMsg RPC
	if err := dec.Decode(buf, &outMsg); err != nil {
		t.Fatalf("decode msg after stream: %v", err)
	}
	if outMsg.Stream {
		t.Fatalf("outMsg should not be stream")
	}
	if string(outMsg.Payload) != "after stream" {
		t.Fatalf("outMsg payload = %q", outMsg.Payload)
	}
}
