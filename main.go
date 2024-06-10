package main

import (
	"filesync/p2p"
	"fmt"
	"log"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransport := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		ShakeHands:    p2p.NOPHandshake,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer: func(p p2p.Peer) error {
			fmt.Println("New peer connected", p)
			return nil
		},
	})

	return NewFileServer(FileServerOpts{
		StorageRoot:       "network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	})
}

func main() {
	s1 := makeServer(":4000")
	s2 := makeServer(":4001", ":4000")

	go func() {
		log.Fatal(s1.Start())
	}()
	go func() {
		log.Fatal(s2.Start())
	}()

	select {}
}
