package websockets

import (
	"fmt"
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
	ErrHijackerNotSatisfied 	 = errors.New("response does not implement hijacker interface")
	ErrBufferNotEmpty 			 = errors.New("cliend sent data before handshake")
)

func (upg *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request) (*Connection, error) {
	if r.Method != "GET" {
		return nil, ErrBadMethod
	}
	if r.Header.Get("connection") != "upgrade" {
		return nil, ErrBadConnectionHeader
	}
	if r.Header.Get("upgrade") != "websocket" {
		return nil, ErrBadUpgradeHeader
	}
	if r.Header.Get("sec-websocket-version") != "13" {
		return nil, ErrBadWebsocketVersionHeader
	}
	key := r.Header.Get("sec-websocket-key")
	if len(key) == 0 {
		return nil, ErrBadWebsocketKeyHeader
	}

	h,ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrHijackerNotSatisfied
	}
	
	netCon,buff,err := h.Hijack()
	if err != nil {
		return nil, fmt.Errorf("failed to hijack: %s",err.Error())
	}

	if buff.Reader.Buffered() > 0 {
		netCon.Close()
		return nil, ErrBufferNotEmpty
	}
	
	con := &Connection{con:netCon}
	con.replyHandshake(key)
	return con, nil
}

