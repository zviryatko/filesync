package p2p

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

// TCPPeer is a struct that represents a remote node over a TCP connection.
type TCPPeer struct {
	// Conn is the underlying TCP connection.
	net.Conn

	// If we dial and retrieve conn => outbound = true.
	// If we accept and retrieve conn => outbound = false.
	outbound bool
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Write(b)
	return err
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
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

// Close closes the network connection.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial dials a remote node.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn)

	return nil
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

	log.Printf("listening on %s\n", t.ListenAddress)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	defer func() {
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
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if errors.Is(err, io.EOF) {
			fmt.Printf("Peer %s closed connection\n", conn.RemoteAddr())
			return
		}
		if err != nil {
			fmt.Printf("Peer %s read error: %s\n", conn.RemoteAddr(), err)
			return
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}

func (t *TCPTransport) ListenAddr() string {
	return t.ListenAddress
}
