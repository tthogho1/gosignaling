package repository

import (
	"errors"

	"gosignaling/model"
)

// Room defines the interface for room repository
type Room interface {
	Get(roomID string) (*model.Room, error)
	Create(r *model.Room) (*model.Room, error)
	Update(r *model.Room) (*model.Room, error)
	Delete(roomID string) error
	GetByClientID(clientID string) (*model.Room, error)
}

var (
	ErrNotFound = errors.New("room not found")
)
