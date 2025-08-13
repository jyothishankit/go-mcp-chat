package server

import (
	"log"
	"net/http"

	"go-mcp-chat/internal/config"
	"go-mcp-chat/internal/hub"
	"go-mcp-chat/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	config *config.Config
	hub    *hub.Hub
	router *gin.Engine
	upgrader websocket.Upgrader
}

func New(cfg *config.Config) *Server {
	server := &Server{
		config: cfg,
		hub:    hub.New(cfg),
		router: gin.Default(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// Serve static files
	s.router.Static("/static", "./static")
	s.router.LoadHTMLGlob("templates/*")

	// API routes
	api := s.router.Group("/api")
	{
		api.GET("/rooms", s.getRooms)
		api.POST("/rooms", s.createRoom)
		api.GET("/rooms/:id", s.getRoom)
		api.DELETE("/rooms/:id", s.deleteRoom)
		api.GET("/stats", s.getStats)
	}

	// WebSocket endpoint
	s.router.GET("/ws", s.handleWebSocket)

	// Serve chat interface
	s.router.GET("/", s.serveChatInterface)
}

func (s *Server) Start() error {
	addr := s.config.Host + ":" + s.config.Port
	return s.router.Run(addr)
}

func (s *Server) getRooms(c *gin.Context) {
	rooms := s.hub.GetRooms()
	
	// Create room summaries
	var roomSummaries []gin.H
	for _, room := range rooms {
		roomSummaries = append(roomSummaries, gin.H{
			"id":           room.ID,
			"name":         room.Name,
			"client_count": room.GetClientCount(),
			"created_at":   room.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    roomSummaries,
	})
}

func (s *Server) createRoom(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Room name is required",
		})
		return
	}

	room := s.hub.CreateRoom(req.Name)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":   room.ID,
			"name": room.Name,
		},
	})
}

func (s *Server) getRoom(c *gin.Context) {
	roomID := c.Param("id")
	room := s.hub.GetRoom(roomID)

	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Room not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           room.ID,
			"name":         room.Name,
			"client_count": room.GetClientCount(),
			"created_at":   room.CreatedAt,
		},
	})
}

func (s *Server) deleteRoom(c *gin.Context) {
	roomID := c.Param("id")
	s.hub.RemoveRoom(roomID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Room deleted",
	})
}

func (s *Server) getStats(c *gin.Context) {
	stats := s.hub.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func (s *Server) handleWebSocket(c *gin.Context) {
	// Get query parameters
	roomID := c.Query("room_id")
	clientName := c.Query("name")
	isGPT := c.Query("gpt") == "true"

	if roomID == "" || clientName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "room_id and name are required",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket first
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	client := models.NewClient(clientName, roomID, conn, isGPT)

	// Handle client in hub (this will create room if it doesn't exist)
	s.hub.HandleClient(client)
}



func (s *Server) serveChatInterface(c *gin.Context) {
	c.HTML(http.StatusOK, "chat.html", gin.H{
		"title": "MCP Chat Server",
	})
}
