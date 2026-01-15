# WebRTC Signaling Server (Go)

A WebRTC signaling server implemented in Go. This implementation is based on the [singo](https://github.com/tockn/singo) library.

## Features

- **Room Management**: Creates rooms when clients connect, enabling protocol exchange between users in the same room
- **Automatic Resource Management**: Automatically deletes room resources when all users leave
- **WebSocket Based**: Bidirectional communication using WebSocket
- **Full Mesh P2P**: Supports full mesh P2P communication between multiple users

## Architecture

```
gosignaling/
├── main.go              # Entry point
├── server.go            # HTTP server configuration
├── handler/
│   └── handler.go       # WebSocket connection and message handling
├── manager/
│   └── room.go          # Room management logic
├── model/
│   ├── room.go          # Room and client models
│   └── message.go       # Message type definitions
└── repository/
    ├── room.go          # Repository interface
    └── mem/
        └── room.go      # In-memory repository implementation
```

## Installation

### Prerequisites

- Go 1.21 or higher

### Setup

```bash
# Clone repository (or create)
cd gosignaling

# Install dependencies
go mod download
```

## Usage

### Starting the Server

```bash
# Start with default settings (0.0.0.0:5000)
go run .

# Start with custom address and port
go run . -addr 127.0.0.1 -port 8080
```

### Command Line Options

- `-addr`: Server address (default: `0.0.0.0`)
- `-port`: Server port (default: `5000`)

### Build

```bash
# Build executable
go build -o signaling-server

# Run
./signaling-server
```

For Windows:

```powershell
go build -o signaling-server.exe
.\signaling-server.exe
```

## WebSocket API

### Endpoint

- `ws://localhost:5000/connect` - WebSocket connection endpoint

### Message Types

#### Client → Server

**1. Join Room**

```json
{
  "type": "join",
  "payload": {
    "room_id": "room123"
  }
}
```

**2. Send SDP Offer**

```json
{
  "type": "offer",
  "payload": {
    "sdp": "v=0\r\no=- ...",
    "client_id": "target_client_id"
  }
}
```

**3. Send SDP Answer**

```json
{
  "type": "answer",
  "payload": {
    "sdp": "v=0\r\no=- ...",
    "client_id": "target_client_id"
  }
}
```

#### Server → Client

**1. Client ID Notification**

```json
{
  "type": "notify-client-id",
  "payload": {
    "client_id": "unique_client_id"
  }
}
```

**2. New Client Notification**

```json
{
  "type": "new-client",
  "payload": {
    "client_id": "new_client_id"
  }
}
```

**3. Client Leave Notification**

```json
{
  "type": "leave-client",
  "payload": {
    "client_id": "leaving_client_id"
  }
}
```

**4. Receive SDP Offer**

```json
{
  "type": "offer",
  "payload": {
    "client_id": "sender_client_id",
    "sdp": "v=0\r\no=- ..."
  }
}
```

**5. Receive SDP Answer**

```json
{
  "type": "answer",
  "payload": {
    "client_id": "sender_client_id",
    "sdp": "v=0\r\no=- ..."
  }
}
```

## Processing Flow

1. **Connection Establishment**

   - Client connects to `/connect` via WebSocket
   - Server generates and notifies a unique client ID

2. **Room Join**

   - Client sends `join` message
   - Room is automatically created if it doesn't exist
   - Existing clients are notified of new participant

3. **WebRTC Connection Establishment**

   - New participant creates Offer and sends to existing clients
   - Existing clients reply with Answer
   - ICE candidate exchange (included in SDP)

4. **Room Leave**
   - Client is automatically removed from room on disconnect
   - Other clients are notified of the departure
   - Resources are automatically deleted when room becomes empty

## Tech Stack

- **Go**: Programming language
- **gorilla/websocket**: WebSocket implementation
- **rs/xid**: Unique ID generation

## License

MIT License

## Reference

This project is implemented based on [tockn/singo](https://github.com/tockn/singo).
