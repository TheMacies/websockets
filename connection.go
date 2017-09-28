package websockets

import (
	"errors"
	"fmt"
	"net"
)

var (
	ErrNoMask = errors.New("message from client must be masked")
)

const (
	DefaultMaxPayloadSize = 10000
)

type Connection interface {
	Write([]byte) error
	GetNextMessage() ([]byte, error)
}

type connection struct {
	con                net.Conn
	rBuff              []byte
	isServer           bool
	maxPayloadPerFrame int
}

func (c *connection) WriteText(payload string) error {
	return c.write([]byte(payload), Text)
}

func (c *connection) WriteBinary(payload []byte) error {
	return c.write(payload, Binary)
}

func (c *connection) write(payload []byte, opCode int) error {
	var err error
	parts := c.dividePayload(payload)
	encodeFrames := make([][]byte, len(parts))
	for i := range parts {
		finFlag := false
		if i == len(parts)-1 {
			finFlag = true
		}
		frameOpCode := Continuation
		if i == 0 {
			frameOpCode = opCode
		}
		enc, err := encodeFrame(&frame{payload: parts[i], OpCode: frameOpCode, finFlag: finFlag, maskUsed: !c.isServer})
		if err != nil {
			return fmt.Errorf("error encoding frame: %s", err.Error())
		}
		encodeFrames[i] = enc
	}

	for i := range encodeFrames {
		_, err := c.con.Write(encodeFrames[i])
		if err != nil {
			return fmt.Errorf("error sending frame: %s", err.Error())
		}
	}
	return nil
}

func (c *connection) GetNextMessage() ([]byte, error) {
	result := []byte{}
	for {
		c.rBuff = c.rBuff[:0]
		_, err := c.con.Read(c.rBuff)
		if err != nil {
			return nil, fmt.Errorf("Error reading from connection: %s", err.Error())
		}
		frame, err := decodeFrame(c.rBuff)
		if err != nil {
			return nil, fmt.Errorf("Error decoding frame: %s", err.Error())
		}

		if frame.OpCode == Ping {
			err := c.sendPong(c.rBuff)
			if err != nil {
				return nil, fmt.Errorf("Error sending pong message")
			}
			continue
		}

		if c.isServer && !frame.maskUsed {
			return nil, ErrNoMask
		}
		result = append(result, frame.payload...)
		if frame.finFlag {
			break
		}
	}
	return result, nil
}

const zeroOpcodeMask = (1<<7 + 1<<6 + 1<<5 + 1<<4) // AND with this masks changes last 4 bits to zeros -> thats where is opCode in the frame

func (c *connection) sendPong(data []byte) error {
	data[0] = (data[0] & zeroOpcodeMask) + Pong
	_, err := c.con.Write(data)
	return err
}

func (c *connection) calculateFrameSize(payload []byte) int {
	size := 2 // FIN , RSV, opCode, mask bit , basic payload len
	if !c.isServer {
		size += 4 // mask must be included
	}
	switch {
	case len(payload) == 127 || len(payload) >= 1<<16:
		size += 8
	case len(payload) == 126 || len(payload) >= 1<<6:
		size += 2
	}
	size += len(payload)
	return size
}

func (c *connection) dividePayload(payload []byte) [][]byte {
	maxPayloadSize := c.maxPayloadPerFrame
	if maxPayloadSize == 0 {
		maxPayloadSize = DefaultMaxPayloadSize
	}
	partsCount := len(payload) / maxPayloadSize
	if partsCount*maxPayloadSize != len(payload) {
		partsCount++
	}
	parts := make([][]byte, 0, partsCount)
	for i := 0; i < partsCount; i++ {
		if i == partsCount-1 {
			parts = append(parts, payload[i*maxPayloadSize:])
		} else {
			parts = append(parts, payload[i*maxPayloadSize:(i+1)*maxPayloadSize])
		}
	}
	return parts
}
