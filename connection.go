package websockets

import (
	"net"
)

type Connection struct {
	con net.Conn
}


func (c *Connection) replyHandshake(challangeKey string) {
	acceptKey := getAcceptKey(challangeKey)
	c.con.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept:" + acceptKey))
}