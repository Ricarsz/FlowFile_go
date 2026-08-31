package main

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/Ricarse/fileFlowSystem/p2p"
)

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("probe port: %v", err)
	}
	addr := l.Addr().String()
	l.Close()
	return addr
}

func TestSmoke_TwoNodesStoreGet(t *testing.T) {
	trA := p2p.NewTcpTransport(freePort(t))
	srvA := NewFileServer(FileServerOpts{
		StorageRoot:       t.TempDir(),
		PathTransformFunc: CASPathTransform,
		Transport:         trA,
	})
	go srvA.Start()

	time.Sleep(100 * time.Millisecond)
	aAddr := trA.Addr()

	trB := p2p.NewTcpTransport(freePort(t))
	srvB := NewFileServer(FileServerOpts{
		StorageRoot:       t.TempDir(),
		PathTransformFunc: CASPathTransform,
		Transport:         trB,
		BootstrapNodes:    []string{aAddr},
	})
	go srvB.Start()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		srvA.peerLock.Lock()
		n := len(srvA.peers)
		srvA.peerLock.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	srvA.peerLock.Lock()
	peersA := len(srvA.peers)
	srvA.peerLock.Unlock()
	if peersA == 0 {
		t.Fatalf("B never connected to A")
	}

	data := []byte("smoke test payload")
	if err := srvA.Store("k1", bytes.NewReader(data)); err != nil {
		t.Fatalf("A.Store: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	t.Logf("after Store push: B.Has(k1)=%v", srvB.Has("k1"))
	if !srvB.Has("k1") {
		t.Fatalf("replica push failed: B does not have k1")
	}

	if err := srvB.Delete("k1"); err != nil {
		t.Fatalf("B.Delete: %v", err)
	}
	if srvB.Has("k1") {
		t.Fatalf("B still has k1 after delete")
	}
	if err := srvB.Get("k1"); err != nil {
		t.Fatalf("B.Get: %v", err)
	}
	if !srvB.Has("k1") {
		t.Fatalf("B does not have k1 after Get")
	}
	_, r, err := srvB.Read("k1")
	if err != nil {
		t.Fatalf("B.Read: %v", err)
	}
	got := new(bytes.Buffer)
	got.ReadFrom(r)
	r.Close()
	if !bytes.Equal(got.Bytes(), data) {
		t.Fatalf("content mismatch: got %q want %q", got.Bytes(), data)
	}
	t.Logf("smoke OK: replica push + broadcast Get both work")
}
