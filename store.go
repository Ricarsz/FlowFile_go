package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"strings"
)

type PathKey struct {
	PathName string
	FileName string
}

type PathTransformFunc func(string) PathKey

type StoreOpts struct{
	Root string
	PathTransformFunc PathTransformFunc
}

func NewStoreOpts(root string)*StoreOpts{
	return &StoreOpts{
		Root:root,
	}
}

type Store struct{
	opts StoreOpts
}
func NewStore(opts StoreOpts)*Store{
	return &Store{opts}
}

func CASPathTransform(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	paths := []string{hashStr[0:5], hashStr[5:10]}
	return PathKey{PathName: strings.Join(paths, "/"),
		FileName: hashStr}
}

func (s *Store) Write(Key string,r io.Reader)(int64,error){
	pathKey:=CASPathTransform(Key)
	absolutePath:=s.opts.Root+"/"+pathKey.PathName
	if err:=os.MkdirAll(absolutePath,0755);err!=nil{
		return 0,err
	}
	fullpath:=absolutePath+"/"+pathKey.FileName
	file,err:=os.Create(fullpath)
	if err!=nil{
		return 0,err
	}
	defer file.Close()
}




