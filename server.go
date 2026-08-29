package main

import (
	"sync"
	"github.com/Ricarse/fileFlowSystem/p2p"
)

type FileServer struct {
	FileServerOpts FileServerOpts
	Store          *Store
	Peers          map[string]p2p.Peer
	PeerLock       sync.Mutex
}

func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		FileServerOpts: opts,
		Store: NewStore(StoreOpts{Root: opts.StorageRoot,
			PathTransformFunc: opts.PathTransformFunc}),
		Peers: make(map[string]p2p.Peer),
	}
}

type FileServerOpts struct {
	PathTransformFunc PathTransformFunc
	StorageRoot       string
	Transport         p2p.Transport
	BootstrapNodes    []string
}

func (s *FileServer)OnPeer(p p2p.Peer)error{
	s.PeerLock.Lock()
	defer s.PeerLock.Unlock()
	s.Peers[p.Conn().RemoteAddr().String()]=p
	return nil
}