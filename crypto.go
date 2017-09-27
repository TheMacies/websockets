package websockets

import (
	"crypto/sha1"
	"encoding/base64"
)

var (
	AcceptHashAppend = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
)
func getAcceptKey(key string) string {
	sha := sha1.Sum(append([]byte(key), AcceptHashAppend...))
	return base64.StdEncoding.EncodeToString(sha[:])
}
