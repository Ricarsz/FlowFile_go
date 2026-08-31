package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type PathKey struct {
	PathName string
	FileName string
}

type StoreOpts struct {
	Root              string
	PathTransformFunc PathTransformFunc
}

type PathTransformFunc func(string) PathKey

type Store struct {
	mu   sync.RWMutex
	opts StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{opts: opts}
}

func (s *Store) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, err := os.Stat(s.fullPath(key))
	if err != nil {
		return false
	}
	return true
}

func (s *Store) Read(key string) (int64, io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	file, err := os.Open(s.fullPath(key))
	if err != nil {
		return 0, nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return 0, nil, err
	}
	return info.Size(), file, nil
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fullPath := s.fullPath(key)
	if err := os.Remove(fullPath); err != nil {
		return err
	}
	dir := filepath.Dir(fullPath)
	for dir != s.opts.Root && dir != "." && dir != "/" {
		if err := os.Remove(dir); err != nil {
			break
		}
		dir = filepath.Dir(dir)
	}
	return nil
}
func CASPathTransform(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	paths := []string{hashStr[0:5], hashStr[5:10]}
	return PathKey{PathName: strings.Join(paths, "/"),
		FileName: hashStr}
}

func (s *Store) Write(key string, r io.Reader) (n int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fullpath := s.fullPath(key)
	dir := filepath.Dir(fullpath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, err
	}
	file, err := os.Create(fullpath)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cerr := file.Close(); err == nil {
			err = cerr
		}
	}()
	n, err = io.Copy(file, r)
	return
}

func (s *Store) fullPath(key string) string {
	var pathKey PathKey
	if s.opts.PathTransformFunc == nil {
		pathKey = CASPathTransform(key)
	} else {
		pathKey = s.opts.PathTransformFunc(key)
	}
	fullpath := filepath.Join(s.opts.Root, pathKey.PathName, pathKey.FileName)
	return fullpath
}
