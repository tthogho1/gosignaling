package manager

import (
	"encoding/json"
	"log"

	"gosignaling/model"
	"gosignaling/repository"
)

// RoomManager manages room operations
type RoomManager struct {
	roomRepo repository.Room
}

// NewRoomManager creates a new room manager
func NewRoomManager(roomRepo repository.Room) *RoomManager {
	return &RoomManager{
		roomRepo: roomRepo,
	}
}

// JoinRoom handles a client joining a room
func (rm *RoomManager) JoinRoom(c *model.Client, roomID string) error {
	room, err := rm.roomRepo.Get(roomID)
	if err == repository.ErrNotFound {
		// Create room if it doesn't exist
		room = model.NewRoom(roomID)
		if _, err := rm.roomRepo.Create(room); err != nil {
			return err
		}
		log.Printf("Created new room: %s", roomID)
	} else if err != nil {
		return err
	}

	// Add client to room
	room.Clients[c.ID] = c
	if _, err := rm.roomRepo.Update(room); err != nil {
		return err
	}

	log.Printf("Client %s joined room %s", c.ID, roomID)

	// Notify other clients in the room
	return rm.notifyNewClient(roomID, c)
}

// LeaveRoom handles a client leaving a room
func (rm *RoomManager) LeaveRoom(c *model.Client) error {
	room, err := rm.roomRepo.GetByClientID(c.ID)
	if err != nil {
		return err
	}

	// Notify other clients before removing
	go rm.notifyLeaveClient(room.ID, c)

	// Remove client from room
	delete(room.Clients, c.ID)

	// If room is empty, delete it
	if len(room.Clients) == 0 {
		if err := rm.roomRepo.Delete(room.ID); err != nil {
			log.Printf("Error deleting empty room %s: %v", room.ID, err)
		} else {
			log.Printf("Deleted empty room: %s", room.ID)
		}
	} else {
		if _, err := rm.roomRepo.Update(room); err != nil {
			return err
		}
	}

	log.Printf("Client %s left room %s", c.ID, room.ID)
	return nil
}

// notifyNewClient notifies all existing clients about a new client
func (rm *RoomManager) notifyNewClient(roomID string, newClient *model.Client) error {
	room, err := rm.roomRepo.Get(roomID)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{"client_id": newClient.ID})
	msg := &model.Message{
		Type:    model.MessageTypeNewClient,
		Payload: payload,
	}

	for _, client := range room.Clients {
		if client.ID != newClient.ID {
			select {
			case client.Send <- msg:
			default:
				log.Printf("Failed to send new client notification to %s", client.ID)
			}
		}
	}

	return nil
}

// notifyLeaveClient notifies all clients about a client leaving
func (rm *RoomManager) notifyLeaveClient(roomID string, leavingClient *model.Client) error {
	room, err := rm.roomRepo.Get(roomID)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{"client_id": leavingClient.ID})
	msg := &model.Message{
		Type:    model.MessageTypeLeaveClient,
		Payload: payload,
	}

	for _, client := range room.Clients {
		if client.ID != leavingClient.ID {
			select {
			case client.Send <- msg:
			default:
				log.Printf("Failed to send leave notification to %s", client.ID)
			}
		}
	}

	return nil
}

// TransferSDPOffer transfers an SDP offer from one client to another
func (rm *RoomManager) TransferSDPOffer(senderClient *model.Client, sdp *model.SDP, targetClientID string) error {
	room, err := rm.roomRepo.GetByClientID(senderClient.ID)
	if err != nil {
		return err
	}

	targetClient, ok := room.Clients[targetClientID]
	if !ok {
		log.Printf("Target client %s not found in room", targetClientID)
		return nil
	}

	payload, _ := json.Marshal(map[string]string{
		"client_id": senderClient.ID,
		"sdp":       sdp.SDP,
	})
	msg := &model.Message{
		Type:    model.MessageTypeSDPOffer,
		Payload: payload,
	}

	select {
	case targetClient.Send <- msg:
	default:
		log.Printf("Failed to send SDP offer to %s", targetClientID)
	}

	return nil
}

// TransferSDPAnswer transfers an SDP answer from one client to another
func (rm *RoomManager) TransferSDPAnswer(senderClient *model.Client, sdp *model.SDP, targetClientID string) error {
	room, err := rm.roomRepo.GetByClientID(senderClient.ID)
	if err != nil {
		return err
	}

	targetClient, ok := room.Clients[targetClientID]
	if !ok {
		log.Printf("Target client %s not found in room", targetClientID)
		return nil
	}

	payload, _ := json.Marshal(map[string]string{
		"client_id": senderClient.ID,
		"sdp":       sdp.SDP,
	})
	msg := &model.Message{
		Type:    model.MessageTypeSDPAnswer,
		Payload: payload,
	}

	select {
	case targetClient.Send <- msg:
	default:
		log.Printf("Failed to send SDP answer to %s", targetClientID)
	}

	return nil
}
