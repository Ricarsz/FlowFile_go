package p2p

import (
	"net"
	"sync"
	"fmt"
)


type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
}

func NewTcpTransport(add string)*TCPTransport{
	return &TCPTransport{
		listenAddress: add,
		peers: make(map[net.Addr]Peer),
	}
}

type TCPPeer struct{
	tcp net.Conn
}

func(p *TCPPeer) Conn()net.Conn{
	return p.tcp
}

func (p *TCPPeer)Close()error{
	return p.tcp.Close()
}

func(t *TCPTransport)Dial(add string)error{
	conn,err:=net.Dial("tcp",add)
	if err != nil{
		return err
	}
	go t.handleConn(conn)
	return nil
}

func(t *TCPTransport)Close()error{
	if t.listener==nil{
		return nil
	}
	return t.listener.Close()
}

func(t *TCPTransport)Addr()string{
	return  t.listenAddress
}

func (t *TCPTransport) ListenAndAccept()error{
	listener,err := net.Listen("tcp",t.listenAddress)
	if err!=nil {
		return err
	}
	t.listener=listener
	for {
		conn,err:=t.listener.Accept()
		if err !=nil {
			continue
		}
		go t.handleConn(conn)
	}
}

func(t *TCPTransport) handleConn(conn net.Conn){
	peer:=&TCPPeer{tcp:conn}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.peers[conn.RemoteAddr()]=peer
	fmt.Println("log")
}
