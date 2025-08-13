class ChatClient {
    constructor() {
        this.socket = null;
        this.currentRoom = null;
        this.userName = '';
        this.isGPT = false;
        this.rooms = [];
        
        this.initializeEventListeners();
        this.loadRooms();
    }

    initializeEventListeners() {
        // Login form
        document.getElementById('loginForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.joinRoom();
        });

        // Message input
        const messageInput = document.getElementById('messageInput');
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Auto-resize textarea
        messageInput.addEventListener('input', () => {
            messageInput.style.height = 'auto';
            messageInput.style.height = Math.min(messageInput.scrollHeight, 100) + 'px';
        });

        // Create room
        document.getElementById('newRoomName').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.createRoom();
            }
        });
    }

    async loadRooms() {
        try {
            const response = await fetch('/api/rooms');
            const data = await response.json();
            
            if (data.success) {
                this.rooms = data.data;
                this.updateRoomsList();
            }
        } catch (error) {
            console.error('Failed to load rooms:', error);
        }
    }

    updateRoomsList() {
        const roomsList = document.getElementById('roomsList');
        roomsList.innerHTML = '';

        this.rooms.forEach(room => {
            const roomElement = document.createElement('div');
            roomElement.className = 'room-item';
            if (this.currentRoom && this.currentRoom.id === room.id) {
                roomElement.classList.add('active');
            }

            roomElement.innerHTML = `
                <div class="room-name">${room.name}</div>
                <div class="room-info">${room.client_count} users • ID: ${room.id}</div>
            `;

            roomElement.addEventListener('click', () => {
                this.selectRoom(room);
            });

            roomsList.appendChild(roomElement);
        });
    }

    async createRoom() {
        const roomNameInput = document.getElementById('newRoomName');
        const roomName = roomNameInput.value.trim();

        if (!roomName) {
            alert('Please enter a room name');
            return;
        }

        try {
            const response = await fetch('/api/rooms', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name: roomName }),
            });

            const data = await response.json();
            
            if (data.success) {
                roomNameInput.value = '';
                await this.loadRooms();
                this.selectRoom(data.data);
            } else {
                alert('Failed to create room: ' + data.message);
            }
        } catch (error) {
            console.error('Failed to create room:', error);
            alert('Failed to create room');
        }
    }

    selectRoom(room) {
        this.currentRoom = room;
        document.getElementById('currentRoomName').textContent = room.name;
        this.updateRoomsList();
        
        // Auto-join if not connected
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            this.joinRoom();
        }
    }

    joinRoom() {
        const userName = document.getElementById('userName').value.trim();
        const roomId = document.getElementById('roomId').value.trim();
        const isGPT = document.getElementById('gptToggle').checked;

        if (!userName || !roomId) {
            alert('Please enter your name and room ID');
            return;
        }

        this.userName = userName;
        this.isGPT = isGPT;

        // Show connecting status
        this.updateConnectionStatus(false);
        document.getElementById('statusText').textContent = 'Connecting...';

        // Connect to WebSocket
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?room_id=${roomId}&name=${encodeURIComponent(userName)}&gpt=${isGPT}`;
        
        this.socket = new WebSocket(wsUrl);

        this.socket.onopen = () => {
            console.log('WebSocket connected');
            this.updateConnectionStatus(true);
            document.getElementById('loginOverlay').classList.add('hidden');
            document.getElementById('chatContainer').classList.remove('hidden');
            
            // Clear messages
            document.getElementById('messagesContainer').innerHTML = '';
            
            // Update room name if it's a new room
            if (!this.currentRoom) {
                document.getElementById('currentRoomName').textContent = `Room ${roomId}`;
            }
        };

        this.socket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.displayMessage(message);
            } catch (error) {
                console.error('Failed to parse message:', error);
            }
        };

        this.socket.onclose = () => {
            console.log('WebSocket disconnected');
            this.updateConnectionStatus(false);
        };

        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateConnectionStatus(false);
            alert('Failed to connect to chat server. Please try again.');
        };
    }

    updateConnectionStatus(connected) {
        const statusIndicator = document.getElementById('statusIndicator');
        const statusText = document.getElementById('statusText');
        const sendButton = document.getElementById('sendButton');

        if (connected) {
            statusIndicator.classList.add('connected');
            statusText.textContent = 'Connected';
            sendButton.disabled = false;
        } else {
            statusIndicator.classList.remove('connected');
            statusText.textContent = 'Disconnected';
            sendButton.disabled = true;
        }
    }

    sendMessage() {
        const messageInput = document.getElementById('messageInput');
        const content = messageInput.value.trim();

        if (!content || !this.socket || this.socket.readyState !== WebSocket.OPEN) {
            return;
        }

        const message = {
            type: 'message',
            content: content,
            room_id: this.currentRoom.id,
            sender: this.userName
        };

        this.socket.send(JSON.stringify(message));
        messageInput.value = '';
        messageInput.style.height = 'auto';
    }

    displayMessage(message) {
        const messagesContainer = document.getElementById('messagesContainer');
        const messageElement = document.createElement('div');
        
        const messageClass = this.getMessageClass(message.type);
        const avatarClass = this.getAvatarClass(message.type);
        const avatarText = this.getAvatarText(message.sender);
        
        const time = new Date(message.timestamp).toLocaleTimeString();

        messageElement.className = `message ${messageClass}`;
        messageElement.innerHTML = `
            <div class="message-avatar ${avatarClass}">${avatarText}</div>
            <div class="message-content">
                <div class="message-header">
                    <span class="message-sender">${message.sender}</span>
                    <span class="message-time">${time}</span>
                </div>
                <div class="message-text">${this.escapeHtml(message.content)}</div>
            </div>
        `;

        messagesContainer.appendChild(messageElement);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    getMessageClass(type) {
        switch (type) {
            case 'system':
            case 'join':
            case 'leave':
                return 'system';
            case 'gpt':
                return 'gpt';
            default:
                return '';
        }
    }

    getAvatarClass(type) {
        switch (type) {
            case 'gpt':
                return 'gpt';
            case 'system':
            case 'join':
            case 'leave':
                return 'system';
            default:
                return '';
        }
    }

    getAvatarText(sender) {
        if (sender === 'GPT Assistant') {
            return 'AI';
        }
        if (sender === 'System') {
            return 'S';
        }
        return sender.charAt(0).toUpperCase();
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize chat client when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.chatClient = new ChatClient();
});

// Global functions for HTML onclick handlers
function createRoom() {
    window.chatClient.createRoom();
}

function sendMessage() {
    window.chatClient.sendMessage();
}
