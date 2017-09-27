package websockets

import (
	"fmt"
	"net"
)

type Connection interface {
	Write([]byte) error
	GetNextMessage() ([]byte, error)
}

type connection struct {
	con   net.Conn
	rBuff []byte
}

func (c *connection) Write([]byte) error {
	return nil
}

func (c *connection) GetNextMessage() ([]byte, error) {
	data := make([]byte, 0, 100)
	for {
		_, err := c.con.Read(c.rBuff)
		if err != nil {
			return nil, fmt.Errorf("Error reading from connection: %s", err.Error())
		}
		frame := DecodeFrame(c.rBuff)
	}

	return nil, nil
}
