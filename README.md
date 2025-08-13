# Go MCP Chat Server

A real-time multi-client chat server built with Go, featuring WebSocket communication and optional GPT integration for AI-powered responses.

## Features

- **Real-time Chat**: WebSocket-based communication for instant messaging
- **Multi-room Support**: Create and join multiple chat rooms
- **GPT Integration**: Optional AI assistant powered by OpenAI GPT
- **Modern UI**: Beautiful, responsive web interface
- **Room Management**: Create, join, and manage chat rooms
- **Client Types**: Support for regular users and GPT clients
- **Message History**: View recent messages when joining a room
- **Connection Status**: Real-time connection status indicators

## Architecture

The application follows a clean architecture pattern with the following components:

- **Models**: Data structures for messages, clients, and rooms
- **Hub**: Central manager for rooms and client connections
- **Server**: HTTP and WebSocket server with REST API
- **GPT Client**: OpenAI integration for AI responses
- **Config**: Environment-based configuration management

## Prerequisites

- Go 1.21 or higher
- OpenAI API key (optional, for GPT functionality)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-mcp-chat
```

2. Install dependencies:
```bash
go mod tidy
```

3. Create environment file:
```bash
cp env.example .env
```

4. Edit `.env` file with your configuration:
```env
# Server Configuration
PORT=8080
HOST=localhost

# OpenAI Configuration (optional)
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_MODEL=gpt-3.5-turbo

# Chat Configuration
MAX_MESSAGE_LENGTH=1000
MAX_CLIENTS_PER_ROOM=50
```

## Running the Server

1. Start the server:
```bash
go run main.go
```

2. Open your browser and navigate to:
```
http://localhost:8080
```

## Usage

### Creating a Room

1. Enter a room name in the "Room name" field
2. Click "Create" or press Enter
3. The room will appear in the rooms list

### Joining a Room

1. Enter your name in the login form
2. Enter the room ID (you can copy it from the rooms list)
3. Optionally check "Connect as GPT Client" to join as an AI assistant
4. Click "Join Room"

### Chatting

- Type your message in the input field
- Press Enter to send (Shift+Enter for new line)
- Messages are broadcast to all users in the room
- GPT will automatically respond to messages if enabled

### GPT Integration

To enable GPT responses:

1. Set your OpenAI API key in the `.env` file
2. GPT will automatically respond to messages in rooms
3. You can also create a GPT client by checking "Connect as GPT Client"

## API Endpoints

### REST API

- `GET /api/rooms` - List all rooms
- `POST /api/rooms` - Create a new room
- `GET /api/rooms/:id` - Get room details
- `DELETE /api/rooms/:id` - Delete a room
- `GET /api/stats` - Get server statistics

### WebSocket

- `GET /ws?room_id=<id>&name=<name>&gpt=<true|false>` - Connect to chat room

## Message Types

- `text` - Regular user messages
- `gpt` - AI-generated responses
- `system` - System messages (join/leave notifications)
- `join` - User joined notification
- `leave` - User left notification
- `error` - Error messages

## Configuration Options

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080 |
| `HOST` | Server host | localhost |
| `OPENAI_API_KEY` | OpenAI API key | (empty) |
| `OPENAI_MODEL` | GPT model to use | gpt-3.5-turbo |
| `MAX_MESSAGE_LENGTH` | Maximum message length | 1000 |
| `MAX_CLIENTS_PER_ROOM` | Maximum clients per room | 50 |

## Development

### Project Structure

```
go-mcp-chat/
├── main.go                 # Application entry point
├── go.mod                  # Go module file
├── env.example             # Environment variables example
├── README.md               # This file
├── internal/
│   ├── config/             # Configuration management
│   ├── models/             # Data models
│   ├── hub/                # Chat hub and room management
│   ├── server/             # HTTP and WebSocket server
│   └── gpt/                # OpenAI GPT integration
├── templates/              # HTML templates
│   └── chat.html           # Main chat interface
└── static/                 # Static assets
    └── js/
        └── chat.js         # Frontend JavaScript
```

### Building

```bash
go build -o mcp-chat main.go
```

### Running Tests

```bash
go test ./...
```

## Security Considerations

- WebSocket connections allow all origins for development
- Consider implementing authentication for production use
- OpenAI API keys should be kept secure
- Input validation is implemented for message length

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Troubleshooting

### Common Issues

1. **WebSocket connection fails**: Check if the server is running and the port is correct
2. **GPT not responding**: Verify your OpenAI API key is set correctly
3. **Room not found**: Make sure you're using the correct room ID
4. **Messages not sending**: Check your internet connection and server status

### Logs

The server logs important events including:
- Room creation and deletion
- Client connections and disconnections
- Message processing
- GPT API errors

Check the console output for detailed information about any issues.
