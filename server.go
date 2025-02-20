package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/tidwall/resp"
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
		connections: make(map[*Connection]struct{}),
		addConnChan: make(chan *Connection),
		delConnChan: make(chan *Connection),
		msgChan:     make(chan Message),
		kv:          NewKV(),
	}
}

func (s *Server) Start() error {

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	l, err := net.Listen(s.Protocol, addr)
	if err != nil {
		return fmt.Errorf("Error connecting.. %s", addr)
	}
	s.listener = l

	go s.connMsgs()

	slog.Info("Redis server running..", "listenAddr", addr)

	s.acceptConn()
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	connection := NewConnection(conn, s.msgChan, s.delConnChan)
	s.addConnChan <- connection
	connection.read()
}

func (s *Server) acceptConn() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			slog.Error("Error accepting new connection", "err", err)
		} else {
			go s.handleConn(conn)
		}
	}
}

func (s *Server) connMsgs() {
	for {
		select {
		case msgChan := <-s.msgChan:
			if err := s.handleMessage(msgChan); err != nil {
				slog.Error("command process error", "err", err)
			}
		case connection := <-s.addConnChan:
			if len(s.connections) > s.Count {
				msg := fmt.Sprintf("Maximum connections %d reached. Cannot accept new connections...", s.Count)
				slog.Info("Maximum connections reached. Cannot accept new connections.", "Max Connections", s.Count)
				connection.Send([]byte(msg))
				connection.Close()
			} else {
				s.connections[connection] = struct{}{}
				slog.Info("new connection added", "remoteAddr", connection.conn.RemoteAddr())
			}
		case connection := <-s.delConnChan:
			connection.Close()
			delete(s.connections, connection)
			slog.Info("connection removed", "remoteAddr", connection.conn.RemoteAddr())
		}
	}
}

func (s *Server) handleMessage(msg Message) error {
	switch v := msg.cmd.(type) {
	case CLIENTLISTCommand:
		var buf bytes.Buffer
		for c := range s.connections {
			buf.WriteString(c.name + "|")
		}
		clients := buf.String()
		if err := resp.NewWriter(msg.connection.conn).WriteString(clients); err != nil {
			return fmt.Errorf("client reply error: %s", err)
		}
	case CLIENTGETNAMECommand:
		name := msg.connection.name
		if err := resp.NewWriter(msg.connection.conn).WriteString(name); err != nil {
			return fmt.Errorf("client reply error: %s", err)
		}
	case CLIENTSETNAMECommand:
		msg.connection.name = string(v.val)
		if err := resp.NewWriter(msg.connection.conn).WriteString("OK"); err != nil {
			return fmt.Errorf("client reply error: %s", err)
		}
	case HELLOCommand:
		reply := map[string][]byte{
			"server":  []byte("redis"),
			"version": []byte("3.0.0"),
			"proto":   []byte("3"),
		}
		if _, err := msg.connection.Send(WriteJson(reply)); err != nil {
			return fmt.Errorf("client reply error: %s", err)
		}
	case QUITCommand:
		msg.connection.Close()
		delete(s.connections, msg.connection)
	case KEYSCommand:
		keys := strings.Join(s.kv.KEYS(), ",")
		if err := resp.NewWriter(msg.connection.conn).WriteString(keys); err != nil {
			return fmt.Errorf("client reply error: %s", err)
		}
	case DELCommand:
		s.kv.DEL(v.val)
		if err := resp.NewWriter(msg.connection.conn).WriteString("OK"); err != nil {
			return fmt.Errorf("client reply error: %s", err)
		}
	case GETCommand:
		val, ok := s.kv.GET(v.key)
		var res string
		if ok {
			res = string(val)
		} else {
			res = string("key not found")
		}
		if err := resp.NewWriter(msg.connection.conn).WriteString(res); err != nil {
			return fmt.Errorf("client reply error: %s", err)
		}
	case SETCommand:
		s.kv.SET(v.key, v.val)
		if err := resp.NewWriter(msg.connection.conn).WriteString("OK"); err != nil {
			return fmt.Errorf("client reply error: %s", err)
		}
	}
	return nil
}
