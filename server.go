package main

import (
	"bytes"
	"io"
	"maps"
	"sync"

	"github.com/Ricarse/fileFlowSystem/p2p"
)

type FileServer struct {
	fileServerOpts FileServerOpts
	store          *Store
	peers          map[string]p2p.Peer
	pendingHeaders map[string]p2p.MessageStoreFile
	peerLock       sync.Mutex
}

func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		fileServerOpts: opts,
		store: NewStore(StoreOpts{Root: opts.StorageRoot,
			PathTransformFunc: opts.PathTransformFunc}),
		peers:          make(map[string]p2p.Peer),
		pendingHeaders: make(map[string]p2p.MessageStoreFile),
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
			hdr, ok := s.pendingHeaders[rpc.From]
			peer := s.peers[rpc.From]
			s.peerLock.Unlock()
			if !ok || peer == nil {
				continue
			}
			reader := io.LimitReader(peer.Conn(), hdr.Size)
			_, err := s.store.Write(hdr.Key, reader)
			if err != nil {
				peer.Close()
				continue
			}
			s.peerLock.Lock()
			delete(s.pendingHeaders, rpc.From)
			s.peerLock.Unlock()
			if err := peer.CloseStream(); err != nil {
				continue
			}
		} else {
			if hdr, ok := rpc.Payload.(p2p.MessageStoreFile); ok {
				s.peerLock.Lock()
				s.pendingHeaders[rpc.From] = hdr
				s.peerLock.Unlock()
			}
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

func (s *FileServer) Start() error {
	s.fileServerOpts.Transport.OnPeer = s.OnPeer
	go s.fileServerOpts.Transport.ListenAndAccept()
	s.bootstrapNetwork()
	s.loop()
	return nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	size := int64(len(b))
	if _, err := s.store.Write(key, bytes.NewReader(b)); err != nil {
		return err
	}
	s.peerLock.Lock()
	peersCopy := maps.Clone(s.peers)
	s.peerLock.Unlock()
	for _, peer := range peersCopy {
		enc := &p2p.DefaultEncoder{}
		hdrRPC := &p2p.RPC{Payload: p2p.MessageStoreFile{Key: key, Size: size}}
		if err := enc.Encode(peer.Conn(), hdrRPC); err != nil {
			continue
		}
		streamRPC := &p2p.RPC{Stream: true, StreamSize: size}
		if err := enc.Encode(peer.Conn(), streamRPC); err != nil {
			continue
		}
		if _, err := io.CopyN(peer.Conn(), bytes.NewReader(b), size); err != nil {
			continue
		}
	}
	return nil
}
