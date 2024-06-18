package main

import (
	"filesync/p2p"
	"fmt"
	"io"
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
	time.Sleep(2 * time.Second)

	//data := bytes.NewReader([]byte("hello world"))
	//err := s2.Store("hello.txt", data)
	//if err != nil {
	//	panic(err)
	//}

	r, err := s2.Get("hello1.txt")
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	select {}
}
