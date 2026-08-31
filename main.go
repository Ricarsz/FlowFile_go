package main

import (
	"log"

	"github.com/Ricarse/fileFlowSystem/handler"
	"github.com/Ricarse/fileFlowSystem/p2p"
	"github.com/labstack/echo/v4"
)

func main() {
	p2pAddr := ":3000"
	httpAddr := ":8080"
	tr := p2p.NewTcpTransport(p2pAddr)

	opts := FileServerOpts{
		StorageRoot:       "./data",
		PathTransformFunc: CASPathTransform,
		Transport:         tr,
		BootstrapNodes:    []string{},
	}

	s := NewFileServer(opts)

	e := echo.New()
	h := handler.New(s)
	h.Register(e)
	go func() {
		log.Printf("http listening on %s", httpAddr)
		if err := e.Start(httpAddr); err != nil {
			log.Printf("http stopped: %v", err)
		}
	}()

	log.Printf("p2p listening on %s, store %s", tr.Addr(), opts.StorageRoot)
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
