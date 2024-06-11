package p2p

import "net"

// Peer is an interface that represents the remote node.
type Peer interface {
	net.Conn
	Send([]byte) error
	Close() error
}

// Transport is anything that handles the communication between the nodes in the
// network. This can be a TCP connection, a UDP connection,
// or any other type of connection.
type Transport interface {
	// ListenAndAccept listens for incoming connections and accepts them.
	ListenAndAccept() error
	// Consume returns a channel that can be used to receive RPC messages.
	Consume() <-chan RPC
	// Close closes the network connection.
	Close() error
	// Dial calls a remote node.
	Dial(string) error
	ListenAddr() string
}
