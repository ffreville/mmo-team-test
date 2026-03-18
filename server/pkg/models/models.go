package models

import (
	"time"
)

type User struct {
	UserID       string     `json:"user_id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	IsBanned     bool       `json:"is_banned"`
}

type Character struct {
	CharacterID string    `json:"character_id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	ClassType   string    `json:"class_type"`
	Level       int       `json:"level"`
	Exp         int64     `json:"exp"`
	CurrentZone string    `json:"current_zone"`
	PositionX   float64   `json:"position_x"`
	PositionY   float64   `json:"position_y"`
	PositionZ   float64   `json:"position_z"`
	Orientation float64   `json:"orientation"`
	IsOnline    bool      `json:"is_online"`
	CreatedAt   time.Time `json:"created_at"`
}

type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}
