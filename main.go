package main

import (
	"log"
)

type Message struct {
	cmd        Command
	connection *Connection
}

func main() {
	cfg := GetConfig()
	server := NewServer(*cfg)
	log.Fatal(server.Start())
}
