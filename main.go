package main

import (
	"log"

	"github.com/Ricarse/fileFlowSystem/p2p"
)

func main() {
	tr := p2p.NewTcpTransport(":3000")

	opts := FileServerOpts{
		StorageRoot:       "./data",
		PathTransformFunc: CASPathTransform,
		Transport:         tr,
		BootstrapNodes:    []string{},
	}

	s := NewFileServer(opts)

	log.Printf("starting file server on %s, store %s", tr.Addr(), opts.StorageRoot)
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
