package hub

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"go-mcp-chat/internal/config"
	"go-mcp-chat/internal/gpt"
	"go-mcp-chat/internal/models"

	"github.com/google/uuid"
)

type Hub struct {
	rooms    map[string]*models.Room
	mu       sync.RWMutex
	gptClient *gpt.Client
	config   *config.Config
}

func New(cfg *config.Config) *Hub {
	return &Hub{
		rooms:     make(map[string]*models.Room),
		gptClient: gpt.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIModel),
		config:    cfg,
	}
}

func (h *Hub) CreateRoom(name string) *models.Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomID := uuid.New().String()
	room := models.NewRoom(roomID, name, h.config.MaxClientsPerRoom)
	h.rooms[roomID] = room

	log.Printf("Created room: %s (ID: %s)", name, roomID)
	return room
}

func (h *Hub) CreateRoomWithID(roomID, name string) *models.Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Use the provided roomID directly, don't generate a new UUID
	room := models.NewRoom(roomID, name, h.config.MaxClientsPerRoom)
	h.rooms[roomID] = room

	log.Printf("Created room: %s (ID: %s)", name, roomID)
	return room
}

func (h *Hub) GetRoom(roomID string) *models.Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}

func (h *Hub) GetRooms() []*models.Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]*models.Room, 0, len(h.rooms))
	for _, room := range h.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (h *Hub) RemoveRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return
	}

	// Disconnect all clients
	for _, client := range room.GetClients() {
		client.Close()
	}

	delete(h.rooms, roomID)
	log.Printf("Removed room: %s", roomID)
}

func (h *Hub) HandleClient(client *models.Client) {
	log.Printf("Handling client %s for room %s", client.Name, client.RoomID)
	
	room := h.GetRoom(client.RoomID)
	if room == nil {
		// Create room if it doesn't exist using the client's room ID
		log.Printf("Room %s not found, creating new room", client.RoomID)
		room = h.CreateRoomWithID(client.RoomID, "Room " + client.RoomID)
		log.Printf("Created new room: %s for client %s", client.RoomID, client.Name)
	} else {
		log.Printf("Found existing room: %s", client.RoomID)
	}

	// Add client to room
	if !room.AddClient(client) {
		log.Printf("Room is full: %s", client.RoomID)
		client.Close()
		return
	}

	log.Printf("Client %s joined room %s", client.Name, client.RoomID)

	// Send recent messages to new client
	recentMessages := room.GetRecentMessages(50)
	for _, msg := range recentMessages {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			continue
		}
		client.SendMessage(data)
	}

	// Handle client messages
	go h.handleClientMessages(client, room)
}

func (h *Hub) handleClientMessages(client *models.Client, room *models.Room) {
	defer func() {
		room.RemoveClient(client.ID)
		log.Printf("Client %s disconnected from room %s", client.Name, client.RoomID)
	}()

	// Start goroutine to send messages to client
	go func() {
		for {
			select {
			case message, ok := <-client.Send:
				if !ok {
					return
				}
				
				err := client.Conn.WriteMessage(1, message)
				if err != nil {
					log.Printf("Error sending message to client: %v", err)
					return
				}
			}
		}
	}()

	// Read messages from client
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client: %v", err)
			break
		}
		
		h.ProcessMessage(client, message)
	}
}

func (h *Hub) ProcessMessage(client *models.Client, messageData []byte) {
	log.Printf("Processing message from client %s in room %s: %s", client.Name, client.RoomID, string(messageData))
	
	var chatReq models.ChatRequest
	if err := json.Unmarshal(messageData, &chatReq); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	room := h.GetRoom(client.RoomID)
	if room == nil {
		log.Printf("Room not found: %s", client.RoomID)
		return
	}

	log.Printf("Processing %s message from %s: %s", chatReq.Type, client.Name, chatReq.Content)

	// Validate message length
	if len(chatReq.Content) > h.config.MaxMessageLength {
		errorMsg := models.NewMessage(models.MessageTypeError, "Message too long", "System", client.RoomID)
		room.AddMessage(errorMsg)
		return
	}

	switch chatReq.Type {
	case "message":
		// Regular text message
		msg := models.NewMessage(models.MessageTypeText, chatReq.Content, client.Name, client.RoomID)
		room.AddMessage(msg)
		log.Printf("Added message to room %s from %s: %s", client.RoomID, client.Name, chatReq.Content)

		// Check if GPT should respond
		if h.gptClient != nil && h.gptClient.IsAvailable() {
			go h.handleGPTResponse(room, chatReq.Content)
		}

	case "gpt_request":
		// Direct GPT request
		if h.gptClient != nil && h.gptClient.IsAvailable() {
			go h.handleGPTResponse(room, chatReq.Content)
		} else {
			errorMsg := models.NewMessage(models.MessageTypeError, "GPT is not available", "System", client.RoomID)
			room.AddMessage(errorMsg)
		}

	default:
		log.Printf("Unknown message type: %s", chatReq.Type)
	}
}

func (h *Hub) handleGPTResponse(room *models.Room, userMessage string) {
	// Get recent conversation for context
	recentMessages := room.GetRecentMessages(10)
	var conversation []string
	
	for _, msg := range recentMessages {
		if msg.Type == models.MessageTypeText {
			conversation = append(conversation, msg.Content)
		}
	}

	// Generate GPT response
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := h.gptClient.GenerateResponse(ctx, conversation, userMessage)
	if err != nil {
		log.Printf("GPT error: %v", err)
		errorMsg := models.NewMessage(models.MessageTypeError, "Sorry, I'm having trouble responding right now.", "GPT", room.ID)
		room.AddMessage(errorMsg)
		return
	}

	// Create GPT message
	gptMsg := models.NewMessage(models.MessageTypeGPT, response, "GPT Assistant", room.ID)
	gptMsg.Metadata["is_ai"] = true
	room.AddMessage(gptMsg)
}

func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalClients := 0
	for _, room := range h.rooms {
		totalClients += room.GetClientCount()
	}

	return map[string]interface{}{
		"total_rooms":    len(h.rooms),
		"total_clients":  totalClients,
		"gpt_available":  h.gptClient != nil && h.gptClient.IsAvailable(),
	}
}
