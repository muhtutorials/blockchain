package network

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

type Peer struct {
	conn     net.Conn
	incoming bool
}

type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	rpcCh      chan RPC
	addPeerCh  chan *Peer
	peersMu    sync.RWMutex
	peers      map[net.Addr]*Peer
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
		rpcCh:      make(chan RPC),
		addPeerCh:  make(chan *Peer),
		peers:      make(map[net.Addr]*Peer),
	}
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln

	fmt.Println("TCP listening on port", t.listenAddr)
	go t.acceptLoop()

	return nil
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("accept error from (%s) connection: %s\n", conn.RemoteAddr(), err)
		}

		fmt.Printf("[%s] new incoming connection: %s\n", t.listenAddr, conn.RemoteAddr())
		t.AddPeer(conn, true)
	}
}

func (t *TCPTransport) AddPeer(conn net.Conn, incoming bool) {
	peer := &Peer{
		conn:     conn,
		incoming: incoming,
	}
	t.peers[peer.conn.RemoteAddr()] = peer
	go func() {
		t.addPeerCh <- peer
	}()
	go t.readLoop(peer)
}

func (t *TCPTransport) readLoop(peer *Peer) {
	buf := make([]byte, 1<<10) // 1024 bytes
	for {
		n, err := peer.conn.Read(buf)
		if err != nil {
			fmt.Printf("read error from (%s) connection: %s", peer.conn.RemoteAddr(), err)
			continue
		}
		msg := buf[:n]
		t.rpcCh <- RPC{
			From:    peer.conn.RemoteAddr(),
			Payload: bytes.NewReader(msg),
		}
	}
}

func (t *TCPTransport) SendMessage(to net.Addr, payload []byte) error {
	peer, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("%s: could not send message to %s", t.listenAddr, to)
	}
	_, err := peer.conn.Write(payload)
	if err != nil {
		return err
	}
	return nil
}

func (t *TCPTransport) Broadcast(payload []byte) error {
	for addr := range t.peers {
		if err := t.SendMessage(addr, payload); err != nil {
			fmt.Printf("error sending message to peer (%s): %s\n", addr, err)
		}
	}
	return nil
}
