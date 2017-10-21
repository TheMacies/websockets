package websockets

import (
	"crypto/sha1"
	"encoding/base64"
	"math/rand"
	"time"
)

var (
	//AcceptKeyHashAppend is RFC defined string that is supposed to be appended to every challange key before hashing it
	AcceptKeyHashAppend = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
)

func getAcceptKey(key string) string {
	sha := sha1.Sum(append([]byte(key), AcceptKeyHashAppend...))
	return base64.StdEncoding.EncodeToString(sha[:])
}

var random = rand.New(rand.NewSource(time.Now().Unix()))
var byteMask = 1<<8 - 1

func generateMask() [4]byte {
	val := random.Int()
	res := [4]byte{}
	for i := 0; i < 4; i++ {
		res[i] = byte(val & byteMask)
		val >>= 8
	}
	return res
}
