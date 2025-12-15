package services

import (
	"encoding/json"
	"log"

	"gosignaling/config"
	"gosignaling/model"

	"github.com/go-redis/redis/v8"
)

// ClusteringService handles Redis Pub/Sub for multi-pod WebRTC signaling
type ClusteringService struct {
	roomManager RoomManagerInterface
}

// RoomManagerInterface defines methods needed from RoomManager
type RoomManagerInterface interface {
	GetClientByID(clientID string) (*model.Client, error)
	GetRoomByClientID(clientID string) (*model.Room, error)
}

// NewClusteringService creates a new clustering service
func NewClusteringService(rm RoomManagerInterface) *ClusteringService {
	return &ClusteringService{
		roomManager: rm,
	}
}

// InitializeRedisSubscriptions subscribes to Redis channels for clustering
func (cs *ClusteringService) InitializeRedisSubscriptions() {
	pubsub := config.Rdb.Subscribe(config.Ctx,
		string(model.RedisMessageTypeSDPOffer),
		string(model.RedisMessageTypeSDPAnswer),
		string(model.RedisMessageTypeIceCandidate),
		string(model.RedisMessageTypeNewClient),
		string(model.RedisMessageTypeLeaveClient),
	)

	log.Println("📡 Subscribed to Redis Pub/Sub channels for WebRTC signaling clustering")

	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			cs.handleRedisMessage(msg)
		}
	}()
}

// handleRedisMessage processes incoming Redis Pub/Sub messages
func (cs *ClusteringService) handleRedisMessage(msg *redis.Message) {
	var redisMsg model.RedisMessage
	if err := json.Unmarshal([]byte(msg.Payload), &redisMsg); err != nil {
		log.Printf("❌ Failed to unmarshal Redis message: %v", err)
		return
	}

	// Get target client (only handle if client is on this pod)
	targetClient, err := cs.roomManager.GetClientByID(redisMsg.TargetClientID)
	if err != nil {
		// Client not on this pod, ignore
		return
	}

	switch msg.Channel {
	case string(model.RedisMessageTypeSDPOffer):
		cs.handleSDPOffer(targetClient, redisMsg)
	case string(model.RedisMessageTypeSDPAnswer):
		cs.handleSDPAnswer(targetClient, redisMsg)
	case string(model.RedisMessageTypeIceCandidate):
		cs.handleIceCandidate(targetClient, redisMsg)
	case string(model.RedisMessageTypeNewClient):
		cs.handleNewClient(targetClient, redisMsg)
	case string(model.RedisMessageTypeLeaveClient):
		cs.handleLeaveClient(targetClient, redisMsg)
	default:
		log.Printf("⚠️ Unknown Redis channel: %s", msg.Channel)
	}
}

// handleSDPOffer handles SDP offer from Redis
func (cs *ClusteringService) handleSDPOffer(targetClient *model.Client, redisMsg model.RedisMessage) {
	msg := &model.Message{
		Type:    model.MessageTypeSDPOffer,
		Payload: redisMsg.Payload,
	}

	select {
	case targetClient.Send <- msg:
		log.Printf("📤 Forwarded SDP offer from Redis to client %s", targetClient.ID)
	default:
		log.Printf("⚠️ Failed to send SDP offer to client %s", targetClient.ID)
	}
}

// handleSDPAnswer handles SDP answer from Redis
func (cs *ClusteringService) handleSDPAnswer(targetClient *model.Client, redisMsg model.RedisMessage) {
	msg := &model.Message{
		Type:    model.MessageTypeSDPAnswer,
		Payload: redisMsg.Payload,
	}

	select {
	case targetClient.Send <- msg:
		log.Printf("📤 Forwarded SDP answer from Redis to client %s", targetClient.ID)
	default:
		log.Printf("⚠️ Failed to send SDP answer to client %s", targetClient.ID)
	}
}

// handleIceCandidate handles ICE candidate from Redis
func (cs *ClusteringService) handleIceCandidate(targetClient *model.Client, redisMsg model.RedisMessage) {
	msg := &model.Message{
		Type:    model.MessageTypeIceCandidate,
		Payload: redisMsg.Payload,
	}

	select {
	case targetClient.Send <- msg:
		log.Printf("📤 Forwarded ICE candidate from Redis to client %s", targetClient.ID)
	default:
		log.Printf("⚠️ Failed to send ICE candidate to client %s", targetClient.ID)
	}
}

// handleNewClient handles new client notification from Redis
func (cs *ClusteringService) handleNewClient(targetClient *model.Client, redisMsg model.RedisMessage) {
	msg := &model.Message{
		Type:    model.MessageTypeNewClient,
		Payload: redisMsg.Payload,
	}

	select {
	case targetClient.Send <- msg:
		log.Printf("📤 Forwarded new client notification from Redis to client %s", targetClient.ID)
	default:
		log.Printf("⚠️ Failed to send new client notification to client %s", targetClient.ID)
	}
}

// handleLeaveClient handles client leave notification from Redis
func (cs *ClusteringService) handleLeaveClient(targetClient *model.Client, redisMsg model.RedisMessage) {
	msg := &model.Message{
		Type:    model.MessageTypeLeaveClient,
		Payload: redisMsg.Payload,
	}

	select {
	case targetClient.Send <- msg:
		log.Printf("📤 Forwarded leave notification from Redis to client %s", targetClient.ID)
	default:
		log.Printf("⚠️ Failed to send leave notification to client %s", targetClient.ID)
	}
}

// PublishToRedis publishes a message to Redis Pub/Sub
func (cs *ClusteringService) PublishToRedis(channel model.RedisMessageType, redisMsg *model.RedisMessage) error {
	msgBytes, err := json.Marshal(redisMsg)
	if err != nil {
		return err
	}

	if err := config.Rdb.Publish(config.Ctx, string(channel), msgBytes).Err(); err != nil {
		return err
	}

	log.Printf("📡 Published message to Redis channel %s", channel)
	return nil
}
