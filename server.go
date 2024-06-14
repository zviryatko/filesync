package main

import (
	"bytes"
	"encoding/gob"
	"filesync/p2p"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
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

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int
}

func (s *FileServer) broadcast(msg *Message) error {
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)

	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. store file on disk.
	// 2. broadcast to all known peers.

	payload := []byte("large file")

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: len(payload),
		},
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(3 * time.Second)

	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	//buf := new(bytes.Buffer)
	//tee := io.TeeReader(r, buf)
	//if err := s.store.writeStream(key, tee); err != nil {
	//	return err
	//}
	//p := &DataMessage{
	//	Key:  key,
	//	Data: buf.Bytes(),
	//}
	//
	//fmt.Println(buf.Bytes())
	//
	//return s.broadcast(&Message{
	//	From:    s.Transport.ListenAddr(),
	//	Payload: p,
	//})
	return nil
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

func (f *FileServer) handleMessage(from string, m *Message) error {
	switch p := m.Payload.(type) {
	case MessageStoreFile:
		return f.handleMessageStoreFile(from, p)
	}
	return nil
}

func (f *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("unknown peer: %s", from)
	}

	if err := f.store.writeStream(msg.Key, io.LimitReader(peer, int64(msg.Size))); err != nil {
		return err
	}

	peer.(*p2p.TCPPeer).Wg.Done()

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
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Printf("failed to decode payload: %s\n", err)
			}
			if err := s.handleMessage(rpc.From.String(), &msg); err != nil {
				log.Printf("failed to handle message: %s\n", err)
			}
		case <-s.quitch:
			return
		}
	}
}

func init() {
	gob.Register(MessageStoreFile{})
}
