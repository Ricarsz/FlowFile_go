package main

import (
	"bytes"
	"io"
	"testing"
)

func TestStore_WriteHasReadDelete(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(StoreOpts{Root: dir})
	key := "hello.txt"
	data := []byte("hello world")
	n, err := s.Write(key, bytes.NewReader(data))
	if err != nil || n != int64(len(data)) {
		t.Fatalf("write %v %d", err, n)
	}
	if !s.Has(key) {
		t.Fatalf("has false")
	}
	size, r, err := s.Read(key)
	if err != nil {
		t.Fatalf("read %v", err)
	}
	got, _ := io.ReadAll(r)
	r.Close()
	if string(got) != string(data) || size != int64(len(data)) {
		t.Fatalf("read got %q size %d", got, size)
	}
	if err := s.Delete(key); err != nil {
		t.Fatalf("delete %v", err)
	}
	if s.Has(key) {
		t.Fatalf("has after delete")
	}
}
