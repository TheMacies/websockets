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
	for {
		_, err := c.con.Read(c.rBuff)
		if err != nil {
			return nil, fmt.Errorf("Error reading from connection: %s", err.Error())
		}
		frame, err := DecodeFrame(c.rBuff)
		if err != nil {
			return nil, fmt.Errorf("Error decoding frame: %s", err.Error())
		}
		fmt.Println(frame.payload)
	}
	return nil, nil
}
