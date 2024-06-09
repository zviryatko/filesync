package p2p

import (
	"fmt"
	"net"
)

// TCPPeer is a struct that represents a remote node over a TCP connection.
type TCPPeer struct {
	conn net.Conn

	// If we dial and retrieve conn => outbound = true.
	// If we accept and retrieve conn => outbound = false.
	outbound bool
}

// Close closes network the connection.
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddress string
	ShakeHands    HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

// Consume returns a channel that can be used to receive RPC messages.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s\n", err)
		_ = conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err := t.ShakeHands(peer); err != nil {
		_ = conn.Close()
		fmt.Printf("handshake error: %s\n", err)
		return
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			_ = conn.Close()
			fmt.Printf("on peer error: %s\n", err)
			return
		}
	}

	// Read loop.
	rpc := RPC{}
	for {
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("TCP read error: %s\n", err)
			return
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
