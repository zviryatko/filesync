package p2p

// HandshakeFunc is a function that is called when a new connection is
// established to check if the connection is valid.
type HandshakeFunc func(Peer) error

func NOPHandshake(Peer) error {
	return nil
}
