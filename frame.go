package websockets

import (
	"encoding/binary"
	"errors"
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

func DecodeFrame(data []byte) (*frame, error) {
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
