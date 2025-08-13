package gpt

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(apiKey, model string) *Client {
	if apiKey == "" {
		return nil
	}
	
	return &Client{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (c *Client) GenerateResponse(ctx context.Context, conversation []string, userMessage string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("GPT client not initialized - API key required")
	}

	// Build conversation history
	var messages []openai.ChatCompletionMessage
	
	// Add system message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful AI assistant in a group chat. Keep your responses concise, friendly, and relevant to the conversation. You can see the recent conversation history to provide context-aware responses.",
	})

	// Add conversation history (last 10 messages to stay within token limits)
	historyLimit := 10
	if len(conversation) > historyLimit {
		conversation = conversation[len(conversation)-historyLimit:]
	}
	
	for _, msg := range conversation {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg,
		})
	}

	// Add current user message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	// Generate response
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: messages,
			MaxTokens: 150,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate GPT response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response generated from GPT")
	}

	response := strings.TrimSpace(resp.Choices[0].Message.Content)
	if response == "" {
		return "I'm here to help! What would you like to discuss?", nil
	}

	return response, nil
}

func (c *Client) IsAvailable() bool {
	return c.client != nil
}
