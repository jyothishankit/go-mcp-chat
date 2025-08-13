package models

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type Room struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Clients   map[string]*Client `json:"clients"`
	Messages  []*Message        `json:"messages"`
	mu        sync.RWMutex      `json:"-"`
	CreatedAt time.Time         `json:"created_at"`
	MaxClients int              `json:"max_clients"`
}

func NewRoom(id, name string, maxClients int) *Room {
	return &Room{
		ID:         id,
		Name:       name,
		Clients:    make(map[string]*Client),
		Messages:   make([]*Message, 0),
		CreatedAt:  time.Now(),
		MaxClients: maxClients,
	}
}

func (r *Room) AddClient(client *Client) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Clients) >= r.MaxClients {
		return false
	}

	// Check if a client with the same name already exists
	for _, existingClient := range r.Clients {
		if existingClient.Name == client.Name {
			// Remove the existing client with the same name
			delete(r.Clients, existingClient.ID)
			existingClient.Close()
		}
	}

	r.Clients[client.ID] = client
	
	// Add join message
	joinMsg := NewMessage(MessageTypeJoin, client.Name+" joined the room", client.Name, r.ID)
	r.Messages = append(r.Messages, joinMsg)
	
	// Broadcast join message
	r.broadcastMessage(joinMsg, nil)
	
	return true
}

func (r *Room) RemoveClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, exists := r.Clients[clientID]
	if !exists {
		return
	}

	delete(r.Clients, clientID)
	client.Close()

	// Add leave message
	leaveMsg := NewMessage(MessageTypeLeave, client.Name+" left the room", client.Name, r.ID)
	r.Messages = append(r.Messages, leaveMsg)
	
	// Broadcast leave message
	r.broadcastMessage(leaveMsg, nil)
}

func (r *Room) AddMessage(message *Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Messages = append(r.Messages, message)
	r.broadcastMessage(message, nil)
}

func (r *Room) broadcastMessage(message *Message, excludeClient *Client) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	log.Printf("Broadcasting message to %d clients in room %s: %s", len(r.Clients), r.ID, message.Content)
	
	for _, client := range r.Clients {
		if excludeClient != nil && client.ID == excludeClient.ID {
			continue
		}
		log.Printf("Sending message to client %s (%s)", client.Name, client.ID)
		client.SendMessage(data)
	}
}

func (r *Room) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Clients)
}

func (r *Room) GetClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	clients := make([]*Client, 0, len(r.Clients))
	for _, client := range r.Clients {
		clients = append(clients, client)
	}
	return clients
}

func (r *Room) GetRecentMessages(limit int) []*Message {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if limit <= 0 || limit > len(r.Messages) {
		limit = len(r.Messages)
	}
	
	start := len(r.Messages) - limit
	return r.Messages[start:]
}
