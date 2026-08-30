package main

import (
	"bytes"
	"io"
	"sync"

	"github.com/Ricarse/fileFlowSystem/p2p"
)

type FileServer struct {
	fileServerOpts FileServerOpts
	store          *Store
	peers          map[string]p2p.Peer
	peerLock       sync.Mutex
}

func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		fileServerOpts: opts,
		store: NewStore(StoreOpts{Root: opts.StorageRoot,
			PathTransformFunc: opts.PathTransformFunc}),
		peers: make(map[string]p2p.Peer),
	}
}

type FileServerOpts struct {
	PathTransformFunc PathTransformFunc
	StorageRoot       string
	Transport         *p2p.TCPTransport
	BootstrapNodes    []string
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.Conn().RemoteAddr().String()] = p
	return nil
}

func (s *FileServer) loop() error {
	for rpc := range s.fileServerOpts.Transport.Consume() {
		if rpc.Stream {
			s.peerLock.Lock()
			peer := s.peers[rpc.From]
			s.peerLock.Unlock()
			if peer == nil {
				continue
			}
			reader := io.LimitReader(peer.Conn(), rpc.StreamSize)
			_, err := s.store.Write(rpc.From, reader)
			if err != nil {
				peer.Close()
				continue
			}
			if err := peer.CloseStream(); err != nil {
				continue
			}
		} else {
			s.store.Write(rpc.From, bytes.NewReader(rpc.Payload))
		}
	}
	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.fileServerOpts.BootstrapNodes {
		if addr == "" {
			continue
		}
		go func(a string) { s.fileServerOpts.Transport.Dial(a) }(addr)
	}
	return nil
}

func (s *FileServer) broadcastNetwork(key string, b []byte) error {
	for _, peer := range s.peers {
		if err := peer.Send(b); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileServer) Start() error {
	s.fileServerOpts.Transport.OnPeer = s.OnPeer
	go s.fileServerOpts.Transport.ListenAndAccept()
	s.bootstrapNetwork()
	s.loop()
	return nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	b, _ := io.ReadAll(r)
	_, err := s.store.Write(key, bytes.NewReader(b))
	if err != nil {
		return err
	}
	if err := s.broadcastNetwork(key, b); err != nil {
		return err
	}
	return nil
}
