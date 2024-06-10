package main

import (
	"bytes"
	"encoding/gob"
	"filesync/p2p"
	"fmt"
	"io"
	"log"
	"sync"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.RWMutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	if len(s.BootstrapNodes) > 0 {
		s.bootstrapNetwork()
	}
	s.loop()
	return nil
}

type Payload struct {
	Key  string
	Data []byte
}

func (s *FileServer) broadcast(p *Payload) error {
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)

	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. store file on disk.
	// 2. broadcast to all known peers.
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	if err := s.store.writeStream(key, tee); err != nil {
		return err
	}
	p := &Payload{
		Key:  key,
		Data: buf.Bytes(),
	}

	fmt.Println(buf.Bytes())

	return s.broadcast(p)
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p

	log.Printf("peer connected: %s\n", p.RemoteAddr())

	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func() {
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("failed to dial %s: %s\n", addr, err)
			}
		}()
	}
	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("FileServer loop quiting...")
		s.Transport.Close()
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&p); err != nil {
				fmt.Printf("failed to decode payload: %s\n", err)
			}
			fmt.Printf("%+v\n", p)
		case <-s.quitch:
			return
		}
	}
}
