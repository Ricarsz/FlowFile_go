package main

import (
	"log"

	"github.com/Ricarse/fileFlowSystem/handler"
	"github.com/Ricarse/fileFlowSystem/p2p"
	"github.com/labstack/echo/v4"
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

	// HTTP (Echo) on :8080
	e := echo.New()
	h := handler.New(s)
	h.Register(e)
	go func() {
		log.Printf("http listening on :8080")
		if err := e.Start(":8080"); err != nil {
			log.Printf("http stopped: %v", err)
		}
	}()

	log.Printf("p2p listening on %s, store %s", tr.Addr(), opts.StorageRoot)
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
