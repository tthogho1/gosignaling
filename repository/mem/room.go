package mem

import (
	"sync"

	"gosignaling/model"
	"gosignaling/repository"
)

type roomRepository struct {
	mutex sync.RWMutex
	rooms map[string]*model.Room
}

// NewRoomRepository creates a new in-memory room repository
func NewRoomRepository() repository.Room {
	return &roomRepository{
		rooms: make(map[string]*model.Room),
	}
}

// Get retrieves a room by ID
func (r *roomRepository) Get(roomID string) (*model.Room, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	room, ok := r.rooms[roomID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return room, nil
}

// Create creates a new room
func (r *roomRepository) Create(room *model.Room) (*model.Room, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.rooms[room.ID] = room
	return room, nil
}

// Update updates an existing room
func (r *roomRepository) Update(room *model.Room) (*model.Room, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.rooms[room.ID]; !ok {
		return nil, repository.ErrNotFound
	}
	r.rooms[room.ID] = room
	return room, nil
}

// Delete removes a room by ID
func (r *roomRepository) Delete(roomID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.rooms[roomID]; !ok {
		return repository.ErrNotFound
	}
	delete(r.rooms, roomID)
	return nil
}

// GetByClientID finds a room by client ID
func (r *roomRepository) GetByClientID(clientID string) (*model.Room, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, room := range r.rooms {
		if _, ok := room.Clients[clientID]; ok {
			return room, nil
		}
	}
	return nil, repository.ErrNotFound
}
