package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

type Client struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	RoomID   string          `json:"room_id"`
	Conn     *websocket.Conn `json:"-"`
	Send     chan []byte     `json:"-"`
	mu       sync.Mutex      `json:"-"`
	JoinedAt time.Time       `json:"joined_at"`
	IsGPT    bool            `json:"is_gpt"`
}

func NewClient(name, roomID string, conn *websocket.Conn, isGPT bool) *Client {
	return &Client{
		ID:       uuid.New().String(),
		Name:     name,
		RoomID:   roomID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		JoinedAt: time.Now(),
		IsGPT:    isGPT,
	}
}

func (c *Client) SendMessage(message []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	select {
	case c.Send <- message:
	default:
		close(c.Send)
	}
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.Conn != nil {
		c.Conn.Close()
	}
	close(c.Send)
}
