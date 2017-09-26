package websockets

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/http"
)

type Upgrader struct {
}

func NewUpgrader() *Upgrader {
	return &Upgrader{}
}

var (
	ErrBadMethod                 = errors.New("bad http method - GET required")
	ErrBadConnectionHeader       = errors.New("bad 'connection' header value - must be 'upgrade'")
	ErrBadUpgradeHeader          = errors.New("bad 'upgrade' header value - must be 'websocket'")
	ErrBadWebsocketVersionHeader = errors.New("bad 'sec-websocket-version' header value - must be '13'")
	ErrBadWebsocketKeyHeader     = errors.New("sec-websocket-key cannot be empty")
)

var (
	AcceptHashAppend = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
)

func (upg *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request) (*Connection, error) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, ErrBadMethod
	}
	if r.Header.Get("connection") != "upgrade" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, ErrBadConnectionHeader
	}
	if r.Header.Get("upgrade") != "websocket" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, ErrBadUpgradeHeader
	}
	if r.Header.Get("sec-websocket-version") != "13" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, ErrBadWebsocketVersionHeader
	}
	key := r.Header.Get("sec-websocket-key")
	if len(key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return nil, ErrBadWebsocketKeyHeader
	}

	key := getAcceptKey(key)

	return nil, nil
}

func getAcceptKey(key string) string {
	sha := sha1.Sum(append([]byte(key), AcceptHashAppend...))
	return base64.StdEncoding.EncodeToString(sha[:])
}
