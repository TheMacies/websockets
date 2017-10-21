package websockets

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Upgrader struct {
	conf *Config
}

type Config struct {
	handshakeTimeout time.Time
}

func NewUpgrader(conf *Config) *Upgrader {
	return &Upgrader{conf: conf}
}

var (
	ErrBadMethod                 = errors.New("bad http method - GET required")
	ErrBadConnectionHeader       = errors.New("bad 'connection' header value - must contain 'upgrade'")
	ErrBadUpgradeHeader          = errors.New("bad 'upgrade' header value - must contain 'websocket'")
	ErrBadWebsocketVersionHeader = errors.New("bad 'sec-websocket-version' header value - must contain '13'")
	ErrBadWebsocketKeyHeader     = errors.New("sec-websocket-key cannot be empty")
	ErrHijackerNotSatisfied      = errors.New("response does not implement hijacker interface")
	ErrBufferNotEmpty            = errors.New("cliend sent data before handshake")
)

var (
	DefaultSubprotocols        = []string{}
	DefautConnectionBufferSize = 10000
)

func (upg *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request) (Connection, error) {
	if r.Method != "GET" {
		return nil, ErrBadMethod
	}
	if !headerContainsValue(r.Header, "connection", "upgrade") {
		return nil, ErrBadConnectionHeader
	}
	if !headerContainsValue(r.Header, "upgrade", "websocket") {
		return nil, ErrBadUpgradeHeader
	}
	if !headerContainsValue(r.Header, "sec-websocket-version", "13") {
		return nil, ErrBadWebsocketVersionHeader
	}
	key := r.Header.Get("sec-websocket-key")
	if len(key) == 0 {
		return nil, ErrBadWebsocketKeyHeader
	}

	h, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrHijackerNotSatisfied
	}

	netCon, buff, err := h.Hijack()
	if err != nil {
		return nil, fmt.Errorf("failed to hijack: %s", err.Error())
	}

	if buff.Reader.Buffered() > 0 {
		netCon.Close()
		return nil, ErrBufferNotEmpty
	}

	handshakeString := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept:" + getAcceptKey(key) + "\r\n"
	// Subprotocols here , but not supported :(((((((
	handshakeString = handshakeString + "\r\n"
	netCon.SetWriteDeadline(upg.conf.handshakeTimeout)
	_, err = netCon.Write([]byte(handshakeString))
	if err != nil {
		netCon.Close()
		return nil, fmt.Errorf("failed to perform handshake: %s", err.Error())
	}

	netCon.SetDeadline(time.Time{})
	return &connection{con: netCon, isServer: false, rBuff: make([]byte, 0, DefautConnectionBufferSize)}, nil
}
