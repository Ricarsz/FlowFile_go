package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"path/filepath"
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
	opts StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{opts}
}

func CASPathTransform(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	paths := []string{hashStr[0:5], hashStr[5:10]}
	return PathKey{PathName: strings.Join(paths, "/"),
		FileName: hashStr}
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	fullpath:=s.fullPath(key)
	dir:=filepath.Dir(fullpath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0,err
	}
	file, err := os.Create(fullpath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	n, err := io.Copy(file, r)
	return n, err
}

func (s *Store)fullPath(key string)string{
	var pathKey PathKey
	if s.opts.PathTransformFunc == nil {
		pathKey = CASPathTransform(key)
	} else {
		pathKey = s.opts.PathTransformFunc(key)
	}
	absolutePath:=s.opts.Root+"/"+pathKey.PathName
	fullpath := absolutePath + "/" + pathKey.FileName
	return fullpath
}