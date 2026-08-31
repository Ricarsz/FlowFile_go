package p2p

import (
	"errors"
	"net"
	"sync"
)

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
	rpcCh         chan RPC
	Decoder       Decoder
	Encoder       Encoder
	HandshakeFunc func(Peer) error
	OnPeer        func(Peer) error
}

func NewTcpTransport(add string) *TCPTransport {
	return &TCPTransport{
		listenAddress: add,
		Decoder:       &DefaultDecoder{},
		Encoder:       &DefaultEncoder{},
		rpcCh:         make(chan RPC, 1024),
		peers:         make(map[net.Addr]Peer),
	}
}

type TCPPeer struct {
	tcp      net.Conn
	outbound bool
	wg       *sync.WaitGroup
}

func NewTCPPeer(tcp net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		tcp:      tcp,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Conn() net.Conn {
	return p.tcp
}

func (p *TCPPeer) Close() error {
	return p.tcp.Close()
}

func (p *TCPPeer) CloseStream() error {
	p.wg.Done()
	return nil
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.tcp.Write(b)
	return err
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
			if errors.Is(err, net.ErrClosed) {
				return err
			}
			continue
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	peer := NewTCPPeer(conn, outbound)
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
	//share memory
	t.peers[conn.RemoteAddr()] = peer
	t.mu.Unlock()
	for {
		var rpc RPC
		rpc.From = conn.RemoteAddr().String()
		if t.Decoder == nil {
			peer.Close()
			return
		}
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			peer.Close()
			return
		}
		if rpc.Stream {
			peer.wg.Add(1)
			t.rpcCh <- rpc
			peer.wg.Wait()
			continue
		}
		t.rpcCh <- rpc
	}
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcCh
}
