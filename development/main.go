package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/tidwall/resp"
)

const defaultListenAddr = ":5001"

type Config struct {
	ListenAddr string
}

type Message struct {
	cmd        Command
	connection *Connection
}

// type Server struct {
// 	Config
// 	connections map[*Connection]bool
// 	ln          net.Listener
// 	addConnChan chan *Connection
// 	delConnChan chan *Connection
// 	quitCh      chan struct{}
// 	msgCh       chan Message
// 	kv          *KV
// }

func NewServer(cfg Config) *Server {

	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}

	return &Server{
		Config:    cfg,
		peers:     make(map[*Connection]bool),
		addPeerCh: make(chan *Connection),
		delPeerCh: make(chan *Connection),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		kv:        NewKV(),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()

	slog.Info("redis server running", "listenAddr", s.ListenAddr)

	return s.acceptLoop()

}

func (s *Server) handleMessage(msg Message) error {
	switch v := msg.cmd.(type) {
	case CLIENTCommand:
		if err := resp.NewWriter(msg.connection.conn).WriteString("OK"); err != nil {
			return err
		}
	case HELLOCommand:
		helloResp := map[string][]byte{
			"server":  []byte("redis"),
			"version": []byte("3.0.0"),
			"proto":   []byte("3"),
		}
		if _, err := msg.connection.Send(respWriteMap(helloResp)); err != nil {
			return fmt.Errorf("peer send error: %s", err)
		}
	case GETCommand:
		val, ok := s.kv.GET(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}
		if err := resp.NewWriter(msg.peer.conn).WriteString(string(val)); err != nil {
			return err
		}
	case SETCommand:
		s.kv.SET(v.key, v.val)
		if err := resp.NewWriter(msg.peer.conn).WriteString("OK"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				slog.Error("raw message error", "err", err)
			}
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			slog.Info("peer connected", "remoteAddr", peer.conn.RemoteAddr())
			s.peers[peer] = true
		case peer := <-s.delPeerCh:
			slog.Info("peer disconnected", "remoteAddr", peer.conn.RemoteAddr())
			delete(s.peers, peer)
		}
	}
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewConnection(conn, s.msgCh, s.delPeerCh)
	s.addPeerCh <- peer
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "error", err, "listenAddress", conn.RemoteAddr())
	}
}

func main() {

	listenAddr := flag.String("listenAddr", defaultListenAddr, "listen address of local redis server")
	flag.Parse()
	server := NewServer(Config{
		ListenAddr: *listenAddr,
	})
	log.Fatal(server.Start())
}
