package protocols

import "net"

// Session represents a proxy connection session
type Session struct {
	Conn     net.Conn
	Node     *Node
	Protocol Protocol
}

// NewSession creates a new session
func NewSession(conn net.Conn, node *Node, protocol Protocol) *Session {
	return &Session{
		Conn:     conn,
		Node:     node,
		Protocol: protocol,
	}
}
