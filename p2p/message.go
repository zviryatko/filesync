package p2p

import "net"

// RPC holds any arbitrary data that is sent between the nodes.
type RPC struct {
	From    net.Addr
	Payload []byte
}
