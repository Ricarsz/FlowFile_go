package p2p

import (
	"net"
	"sync"
)

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
	rpcCh         chan RPC
	HandshakeFunc func(Peer) error
	OnPeer        func(Peer) error
}

func NewTcpTransport(add string) *TCPTransport {
	return &TCPTransport{
		listenAddress: add,
		rpcCh:         make(chan RPC, 1024),
		peers:         make(map[net.Addr]Peer),
	}
}

type TCPPeer struct {
	tcp      net.Conn
	outbound bool
}

func NewTCPPeer(tcp net.Conn) *TCPPeer {
	return &TCPPeer{
		tcp: tcp,
	}
}

func (p *TCPPeer) Conn() net.Conn {
	return p.tcp
}

func (p *TCPPeer) Close() error {
	return p.tcp.Close()
}

func (t *TCPTransport) Dial(add string) error {
	conn, err := net.Dial("tcp", add)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}

func (t *TCPTransport) Close() error {
	if t.listener == nil {
		return nil
	}
	return t.listener.Close()
}

func (t *TCPTransport) Addr() string {
	return t.listenAddress
}

func (t *TCPTransport) ListenAndAccept() error {
	listener, err := net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}
	t.listener = listener
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			continue
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	peer := &TCPPeer{tcp: conn, outbound: outbound}
	if t.HandshakeFunc != nil {
		if err := t.HandshakeFunc(peer); err != nil {
			peer.Close()
			return
		}
	}
	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			peer.Close()
			return
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.peers[conn.RemoteAddr()] = peer
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			peer.Close()
			return
		}
		rpc := RPC{From: conn.RemoteAddr(), Payload: buffer[:n]}
		t.rpcCh <- rpc
	}
}

func (t *TCPTransport) Cosume() <-chan RPC {
	return t.rpcCh
}
