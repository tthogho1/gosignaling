package model

import "github.com/rs/xid"

// Room represents a WebRTC signaling room
type Room struct {
	ID      string
	Name    string
	Clients map[string]*Client
}

// NewRoom creates a new room with the given name
func NewRoom(name string) *Room {
	return &Room{
		ID:      name,
		Name:    name,
		Clients: make(map[string]*Client),
	}
}

// Client represents a connected WebRTC client
type Client struct {
	ID   string
	Name string
	Send chan *Message
}

// NewClient creates a new client with a unique ID
func NewClient(name string) *Client {
	return &Client{
		ID:   xid.New().String(),
		Name: name,
		Send: make(chan *Message, 16),
	}
}
