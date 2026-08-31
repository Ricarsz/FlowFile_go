package p2p

import (
	"bytes"
	"testing"
)

func TestEncodeDecode_MessageStoreFile(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := &DefaultEncoder{}
	dec := &DefaultDecoder{}
	hdr := MessageStoreFile{Key: "a.txt", Size: 11}
	rpc := &RPC{Payload: hdr}
	if err := enc.Encode(buf, rpc); err != nil {
		t.Fatalf("encode: %v", err)
	}
	var out RPC
	if err := dec.Decode(buf, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Stream {
		t.Fatalf("should not be stream")
	}
	got, ok := out.Payload.(MessageStoreFile)
	if !ok {
		t.Fatalf("payload type %T", out.Payload)
	}
	if got.Key != hdr.Key || got.Size != hdr.Size {
		t.Fatalf("got %+v want %+v", got, hdr)
	}
}

func TestEncodeDecode_Stream(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := &DefaultEncoder{}
	dec := &DefaultDecoder{}
	rpc := &RPC{Stream: true, StreamSize: 1024}
	if err := enc.Encode(buf, rpc); err != nil {
		t.Fatalf("encode: %v", err)
	}
	var out RPC
	if err := dec.Decode(buf, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.Stream || out.StreamSize != 1024 {
		t.Fatalf("stream %+v", out)
	}
}

func TestEncodeDecode_Sticky(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := &DefaultEncoder{}
	dec := &DefaultDecoder{}
	hdr := MessageStoreFile{Key: "a", Size: 5}
	enc.Encode(buf, &RPC{Payload: hdr})
	enc.Encode(buf, &RPC{Stream: true, StreamSize: 5})
	var r1, r2 RPC
	dec.Decode(buf, &r1)
	dec.Decode(buf, &r2)
	if r1.Stream || r2.Stream == false {
		t.Fatalf("sticky failed r1 %v r2 %v", r1.Stream, r2.Stream)
	}
}
