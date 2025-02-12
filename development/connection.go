package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/tidwall/resp"
)

const CONN_CLOSING_MSG = "SERVER CLOSING CONNECTION..."
const CONN_CLOSING_FAIL = "SERVER CONNECTION CLOSE REQUEST FAILED."

type Connection struct {
	conn    net.Conn
	name    string
	msgChan chan Message
	delChan chan *Connection
}

func (c *Connection) Send(msg []byte) (int, error) {
	return c.conn.Write(msg)
}

func (c *Connection) SetName(name string) {
	c.name = name
}

func (c *Connection) Close() {
	c.Send([]byte(CONN_CLOSING_MSG))
	if err := c.conn.Close(); err != nil {
		c.Send([]byte(CONN_CLOSING_FAIL))
	}
}

func NewConnection(conn net.Conn, msgChan chan Message, delChan chan *Connection) *Connection {
	return &Connection{
		conn:    conn,
		name:    uuid.NewString(),
		msgChan: msgChan,
		delChan: delChan,
	}
}

func (c *Connection) read() {
	r := resp.NewReader(c.conn)
	for {
		v, _, err := r.ReadValue()
		if err == io.EOF {
			c.delChan <- c
			return
		}
		if err != nil {
			log.Fatal(err)
			errMsg := fmt.Sprintf("Error encountered reading next value - Error=%s", err)
			c.Send([]byte(errMsg))
			continue
		}

		cmd, err := parseVal(v)
		if err != nil {
			fmt.Println(err.Error())
			c.Send([]byte(err.Error()))
			continue
		}

		c.msgChan <- Message{
			cmd:        cmd,
			connection: c,
		}
	}
}

func parseVal(val resp.Value) (Command, error) {

	if val.Type() != resp.Array {
		errMsg := fmt.Sprintf("Unsupported data type - %v received. Only %v Array format is supported.", val.Type(), resp.Array)
		return nil, errors.New(errMsg)
	}

	var cmd Command
	var err error
	var cmd1Str, cmd2Str, dualStr string

	if len(val.Array()) >= 1 {
		cmd1Str = val.Array()[0].String()
	}
	if len(val.Array()) >= 2 {
		cmd2Str = val.Array()[1].String()
	}

	dualStr = strings.TrimSpace(strings.Join([]string{cmd1Str, cmd2Str}, " "))

	switch cmd1Str {
	case GET:
		arg1 := val.Array()[1].String()
		cmd = GETCommand{
			key: arg1,
		}
	case SET:
		arg1 := val.Array()[1].String()
		arg2 := val.Array()[2].Bytes()
		cmd = SETCommand{
			key: arg1,
			val: arg2,
		}
	case DEL:
		keys := val.Array()[1:]
		val := make([]string, 0, len(keys))
		for _, v := range keys {
			val = append(val, v.String())
		}
		cmd = DELCommand{
			val: val,
		}
	case KEYS:
		cmd = KEYSCommand{}
	case QUIT:
		cmd = QUITCommand{}
	case HELLO:
		cmd = HELLOCommand{}
	default:
		switch dualStr {
		case CLIENTSETNAME:
			arg1 := val.Array()[2].Bytes()
			cmd = CLIENTSETNAMECommand{
				val: arg1,
			}
		case CLIENTGETNAME:
			cmd = CLIENTGETNAMECommand{}
		case CLIENTLIST:
			cmd = CLIENTLISTCommand{}
		default:
			errMsg := fmt.Sprintf("Unsupported command %v or %v received.", cmd1Str, dualStr)
			err = errors.New(errMsg)
		}
	}
	return cmd, err
}
