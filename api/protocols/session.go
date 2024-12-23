package protocols

import (
	"net"
	"singdns/api/models"
)

// Session represents a proxy connection session
type Session struct {
	Conn     net.Conn
	Node     *models.Node
	Protocol Protocol
}

// NewSession creates a new session
func NewSession(conn net.Conn, node *models.Node, protocol Protocol) *Session {
	return &Session{
		Conn:     conn,
		Node:     node,
		Protocol: protocol,
	}
}
