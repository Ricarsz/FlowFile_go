package main

import (
	"log"
	"time"

	"github.com/Ricarse/fileFlowSystem/p2p"
)

func main() {
	tr1 := p2p.NewTcpTransport(":3000")
	go func() { log.Fatal(tr1.ListenAndAccept()) }()

	tr2 := p2p.NewTcpTransport(":4000")
	go func() { log.Fatal(tr2.ListenAndAccept()) }()

	time.Sleep(time.Second) // 为什么要 sleep？如果不 sleep 直接 Dial 会怎样？

	// tr2 主动拨 tr1
	if err := tr2.Dial(":3000"); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second) // 等 handleConn 打印 log
	log.Println("test done")
}
