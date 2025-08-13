package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host              string
	Port              string
	OpenAIAPIKey      string
	OpenAIModel       string
	MaxMessageLength  int
	MaxClientsPerRoom int
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	port := getEnv("PORT", "8080")
	host := getEnv("HOST", "localhost")
	openAIKey := getEnv("OPENAI_API_KEY", "")
	openAIModel := getEnv("OPENAI_MODEL", "gpt-3.5-turbo")
	
	maxMessageLength, _ := strconv.Atoi(getEnv("MAX_MESSAGE_LENGTH", "1000"))
	maxClientsPerRoom, _ := strconv.Atoi(getEnv("MAX_CLIENTS_PER_ROOM", "50"))

	return &Config{
		Host:              host,
		Port:              port,
		OpenAIAPIKey:      openAIKey,
		OpenAIModel:       openAIModel,
		MaxMessageLength:  maxMessageLength,
		MaxClientsPerRoom: maxClientsPerRoom,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
