package models

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeSystem   MessageType = "system"
	MessageTypeGPT      MessageType = "gpt"
	MessageTypeJoin     MessageType = "join"
	MessageTypeLeave    MessageType = "leave"
	MessageTypeError    MessageType = "error"
)

type Message struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	Sender    string      `json:"sender"`
	RoomID    string      `json:"room_id"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func NewMessage(msgType MessageType, content, sender, roomID string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Type:      msgType,
		Content:   content,
		Sender:    sender,
		RoomID:    roomID,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

type ChatRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	RoomID  string `json:"room_id"`
	Sender  string `json:"sender"`
}

type ChatResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
