package main

import (
	"bytes"
	"filesync/p2p"
	"fmt"
	"time"
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

	s := NewFileServer(FileServerOpts{
		StorageRoot:       "network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	})

	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":4000")
	s2 := makeServer(":4001", ":4000")

	go s1.Start()
	time.Sleep(1 * time.Second)
	go s2.Start()
	time.Sleep(1 * time.Second)
	data := bytes.NewReader([]byte("hello world"))
	_ = s2.StoreData("hello.txt", data)

	select {}
}
