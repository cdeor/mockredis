package main

import (
	"log/slog"
	"net"
)

type Server struct {
	RedisConfig
	connections map[*Connection]struct{}
	listener    net.Listener
	addConnChan chan *Connection
	delConnChan chan *Connection
	msgChan     chan Message
	kv          *KV
}

func NewServer(cfg RedisConfig) *Server {
	return &Server{
		RedisConfig: cfg,
		connections: make(map[*Connection]bool),
		addConnCh:   make(chan *Connection),
		delConnCh:   make(chan *Connection),
		msgChan:     make(chan Message),
		kv:          NewKV(),
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
