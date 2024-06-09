package main

import (
	"filesync/p2p"
	"log"
)

func OnPeer(peer p2p.Peer) error {
	log.Printf("New peer: %v", peer)
	return nil
}

func main() {
	tr := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddress: ":4000",
		ShakeHands:    p2p.NOPHandshake,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	})

	go func() {
		for {
			msg := <-tr.Consume()
			log.Printf("Received message: %+v", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
