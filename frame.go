package websockets

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type frame struct {
	finFlag  bool //Flag set if this frame is the last one in a message
	OpCode   byte
	payload  []byte
	maskUsed bool
	mask     []byte
}

var (
	ErrInvalidMessage = errors.New("got invalid message")
)

//OpCodes
const (
	Continuation = byte(0x0)
	Text         = byte(0x1)
	Binary       = byte(0x2)
	Ping         = byte(0x9)
	Pong         = byte(0xA)
)

func encodeFrame(fr *frame) ([]byte, error) {
	frameSize := calculateFrameSize(fr.payload, fr.maskUsed)
	if frameSize >= 1<<63 {
		return nil, fmt.Errorf("payload to big, perhaps you should set lower payload per frame size limit")
	}
	encoded := make([]byte, frameSize)
	if fr.finFlag {
		encoded[0] = 1 << 7
	}
	encoded[0] += fr.OpCode
	if fr.maskUsed {
		encoded[1] = 1 << 7
	}
	payloadLen := len(fr.payload)
	currentByte := 2
	switch {
	case payloadLen < 126:
		encoded[1] += byte(payloadLen)
		currentByte++
	case payloadLen == 126 || payloadLen < 1<<16:
		encoded[1] += 126
		encoded[2] += byte(payloadLen >> 8)
		encoded[3] = byte(payloadLen & (1<<8 - 1))
		currentByte += 3
	default:
		encoded[1] += 127
		for i := 2; i <= 9; i++ {
			encoded[i] = byte((payloadLen >> (56 - (uint(i)-2)*8)) & (1<<8 - 1))
		}
		currentByte += 9
	}
	if fr.maskUsed {
		mask := generateMask()
		for i := 0; i < len(mask); i++ {
			encoded[currentByte] = mask[i]
			currentByte++
		}
	}
	copy(encoded[currentByte:], fr.payload)
	return encoded, nil
}

func decodeFrame(data []byte) (*frame, error) {
	if len(data) < 2 {
		return nil, ErrInvalidMessage
	}

	currentBytesRead := 2
	fr := &frame{}
	fr.finFlag = data[0]&0x80 != 0

	fr.OpCode = data[0] & 0xF
	fr.maskUsed = data[1]&0x80 != 0
	payloadLen := uint64(data[1] & 0x7F)

	switch payloadLen {
	case 0x40:
		if len(data) < currentBytesRead+8 {
			return nil, ErrInvalidMessage
		}
		if data[2]&0x80 != 0 {
			return nil, ErrInvalidMessage
		}
		payloadLen = binary.BigEndian.Uint64((data[currentBytesRead : currentBytesRead+8]))
		currentBytesRead += 2
	case 0x3F:
		if len(data) < currentBytesRead+2 {
			return nil, ErrInvalidMessage
		}
		payloadLen = binary.BigEndian.Uint64((data[currentBytesRead : currentBytesRead+2]))
		currentBytesRead += 2
	}
	if fr.maskUsed {
		fr.mask = make([]byte, 4)
		if len(data) < currentBytesRead+4 {
			return nil, ErrInvalidMessage
		}
		copy(fr.mask, data[currentBytesRead:currentBytesRead+4])
		currentBytesRead += 4
	}
	if uint64(len(data)) < uint64(currentBytesRead)+payloadLen {
		return nil, ErrInvalidMessage
	}
	fr.payload = make([]byte, payloadLen)
	currUint := uint64(currentBytesRead)
	for i := uint64(0); i < payloadLen; i++ {
		fr.payload[i] = data[currUint+i]
		if fr.maskUsed {
			fr.payload[i] = fr.payload[i] ^ fr.mask[i%4]
		}
	}
	return fr, nil
}

func calculateFrameSize(payload []byte, addMask bool) uint64 {
	size := uint64(2) // FIN , RSV, opCode, mask bit , basic payload len
	if addMask {
		size += 4 // mask must be included
	}
	switch {
	case len(payload) == 127 || len(payload) >= 1<<16:
		size += 8
	case len(payload) == 126 || len(payload) >= 1<<6:
		size += 2
	}
	size += uint64(len(payload))
	return size
}
