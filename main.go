package main

import (
	"log"

	"go-mcp-chat/internal/server"
	"go-mcp-chat/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create and start server
	srv := server.New(cfg)
	
	log.Printf("Starting MCP Chat Server on %s:%s", cfg.Host, cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
