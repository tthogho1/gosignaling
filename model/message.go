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
