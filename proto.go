package main

import (
	"bytes"
	"fmt"

	"github.com/tidwall/resp"
)

const (
	SET           = "SET"
	GET           = "GET"
	DEL           = "DEL"
	KEYS          = "KEYS"
	QUIT          = "QUIT"
	HELLO         = "HELLO"
	CLIENTLIST    = "CLIENT LIST"
	CLIENTGETNAME = "CLIENT GETNAME"
	CLIENTSETNAME = "CLIENT SETNAME"
)

type Command interface{}

type SETCommand struct {
	key string
	val []byte
}

type GETCommand struct {
	key string
}

type DELCommand struct {
	val []string
}

type KEYSCommand struct{}

type QUITCommand struct{}

type HELLOCommand struct{}

type CLIENTLISTCommand struct{}

type CLIENTGETNAMECommand struct{}

type CLIENTSETNAMECommand struct {
	val []byte
}

func WriteJson(m map[string][]byte) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("%" + fmt.Sprintf("%d\r\n", len(m)))
	wr := resp.NewWriter(buf)
	for k, v := range m {
		wr.WriteString(k + ":")
		wr.WriteBytes(v)
	}
	return buf.Bytes()
}
