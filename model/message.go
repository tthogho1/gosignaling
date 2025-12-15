package model

import "encoding/json"

// MessageType defines the type of signaling message
type MessageType string

const (
	MessageTypeNotifyClientID MessageType = "notify-client-id"
	MessageTypeNewClient      MessageType = "new-client"
	MessageTypeLeaveClient    MessageType = "leave-client"
	MessageTypeSDPOffer       MessageType = "offer"
	MessageTypeSDPAnswer      MessageType = "answer"
	MessageTypeIceCandidate   MessageType = "ice-candidate"
	MessageTypeError          MessageType = "error"
)

// Message represents a signaling message
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SDP represents WebRTC Session Description Protocol data
type SDP struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

// IceCandidate represents WebRTC ICE candidate data
type IceCandidate struct {
	Candidate      string  `json:"candidate"`
	SdpMid         *string `json:"sdpMid,omitempty"`
	SdpMLineIndex  *uint16 `json:"sdpMLineIndex,omitempty"`
	ClientID       string  `json:"client_id"`
}

// RedisMessageType defines the type of Redis Pub/Sub message for clustering
type RedisMessageType string

const (
	RedisMessageTypeSDPOffer      RedisMessageType = "webrtc:offer"
	RedisMessageTypeSDPAnswer     RedisMessageType = "webrtc:answer"
	RedisMessageTypeIceCandidate  RedisMessageType = "webrtc:ice"
	RedisMessageTypeNewClient     RedisMessageType = "webrtc:new_client"
	RedisMessageTypeLeaveClient   RedisMessageType = "webrtc:leave_client"
)

// RedisMessage represents a message sent through Redis Pub/Sub
type RedisMessage struct {
	Type           RedisMessageType `json:"type"`
	SenderClientID string           `json:"sender_client_id"`
	TargetClientID string           `json:"target_client_id"`
	RoomID         string           `json:"room_id,omitempty"`
	Payload        json.RawMessage  `json:"payload"`
}
