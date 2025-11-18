package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gosignaling/manager"
	"gosignaling/model"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// Handler handles WebSocket connections
type Handler struct {
	manager *manager.RoomManager
}

// NewHandler creates a new handler
func NewHandler(mgr *manager.RoomManager) *Handler {
	return &Handler{
		manager: mgr,
	}
}

// CreateConnection handles WebSocket connection establishment
func (h *Handler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := model.NewClient("user")
	ctx := context.Background()

	// Start goroutines for sending and receiving messages
	go h.HandleSendMessage(ctx, client, conn)
	go h.HandleReceiveMessage(client, conn)

	// Send client ID to the newly connected client
	payload, _ := json.Marshal(map[string]string{"client_id": client.ID})
	msg := &model.Message{
		Type:    model.MessageTypeNotifyClientID,
		Payload: payload,
	}
	msgBytes, _ := json.Marshal(msg)
	sendMessage(conn, msgBytes)

	log.Printf("New client connected: %s", client.ID)
}

// HandleSendMessage handles sending messages to a client
func (h *Handler) HandleSendMessage(ctx context.Context, c *model.Client, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
		h.manager.LeaveRoom(c)
		log.Printf("Client disconnected: %s", c.ID)
	}()

	for {
		select {
		case msg := <-c.Send:
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				return
			}
			if err := sendMessage(conn, msgBytes); err != nil {
				log.Printf("Failed to send message: %v", err)
				return
			}
		case <-ticker.C:
			// Send ping to keep connection alive
			if err := sendMessage(conn, []byte(`{"type":"ping"}`)); err != nil {
				log.Printf("Ping failed, client disconnected: %s", c.ID)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// HandleReceiveMessage handles receiving messages from a client
func (h *Handler) HandleReceiveMessage(c *model.Client, conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %v", c.ID, err)
			return
		}

		var req ReceiveMessage
		if err := json.Unmarshal(msgBytes, &req); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		var resp *model.Message
		switch req.Type {
		case "join":
			resp = h.handleJoinRoom(c, req.Payload)
		case "offer":
			resp = h.handleSDPOffer(c, req.Payload)
		case "answer":
			resp = h.handleSDPAnswer(c, req.Payload)
		default:
			log.Printf("Unknown message type: %s", req.Type)
		}

		if resp != nil {
			respBytes, _ := json.Marshal(resp)
			sendMessage(conn, respBytes)
		}
	}
}

// ReceiveMessage represents an incoming message
type ReceiveMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// JoinRoomPayload represents the payload for joining a room
type JoinRoomPayload struct {
	RoomID string `json:"room_id"`
}

func (h *Handler) handleJoinRoom(c *model.Client, payload json.RawMessage) *model.Message {
	var joinPayload JoinRoomPayload
	if err := json.Unmarshal(payload, &joinPayload); err != nil {
		log.Printf("Failed to unmarshal join room payload: %v", err)
		return &model.Message{
			Type:    model.MessageTypeError,
			Payload: []byte(`{"error":"invalid payload"}`),
		}
	}

	if err := h.manager.JoinRoom(c, joinPayload.RoomID); err != nil {
		log.Printf("Failed to join room: %v", err)
		return &model.Message{
			Type:    model.MessageTypeError,
			Payload: []byte(`{"error":"failed to join room"}`),
		}
	}

	return nil
}

// SDPOfferPayload represents the payload for an SDP offer
type SDPOfferPayload struct {
	SDP      string `json:"sdp"`
	ClientID string `json:"client_id"`
}

func (h *Handler) handleSDPOffer(c *model.Client, payload json.RawMessage) *model.Message {
	var offerPayload SDPOfferPayload
	if err := json.Unmarshal(payload, &offerPayload); err != nil {
		log.Printf("Failed to unmarshal SDP offer payload: %v", err)
		return &model.Message{
			Type:    model.MessageTypeError,
			Payload: []byte(`{"error":"invalid payload"}`),
		}
	}

	sdp := &model.SDP{
		Type: "offer",
		SDP:  offerPayload.SDP,
	}

	if err := h.manager.TransferSDPOffer(c, sdp, offerPayload.ClientID); err != nil {
		log.Printf("Failed to transfer SDP offer: %v", err)
		return &model.Message{
			Type:    model.MessageTypeError,
			Payload: []byte(`{"error":"failed to transfer offer"}`),
		}
	}

	return nil
}

// SDPAnswerPayload represents the payload for an SDP answer
type SDPAnswerPayload struct {
	SDP      string `json:"sdp"`
	ClientID string `json:"client_id"`
}

func (h *Handler) handleSDPAnswer(c *model.Client, payload json.RawMessage) *model.Message {
	var answerPayload SDPAnswerPayload
	if err := json.Unmarshal(payload, &answerPayload); err != nil {
		log.Printf("Failed to unmarshal SDP answer payload: %v", err)
		return &model.Message{
			Type:    model.MessageTypeError,
			Payload: []byte(`{"error":"invalid payload"}`),
		}
	}

	sdp := &model.SDP{
		Type: "answer",
		SDP:  answerPayload.SDP,
	}

	if err := h.manager.TransferSDPAnswer(c, sdp, answerPayload.ClientID); err != nil {
		log.Printf("Failed to transfer SDP answer: %v", err)
		return &model.Message{
			Type:    model.MessageTypeError,
			Payload: []byte(`{"error":"failed to transfer answer"}`),
		}
	}

	return nil
}

func sendMessage(conn *websocket.Conn, msg []byte) error {
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	return w.Close()
}
