package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTCPTransport(t *testing.T) {
	listenAddr := ":4000"
	tr := NewTCPTransport(TCPTransportOpts{
		ListenAddress: listenAddr,
		ShakeHands:    NOPHandshake,
		Decoder:       DefaultDecoder{},
	})
	assert.Equal(t, listenAddr, tr.ListenAddress)
	assert.Nil(t, tr.ListenAndAccept())
}
