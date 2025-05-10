package model

import "time"

type Message struct {
	ID        string    `json:"id"`
	Name      string    `json:"name" validate:"required"`
	Email     string    `json:"email" validate:"omitempty,email"`
	Message   string    `json:"message" validate:"required"`
	CreatedAt time.Time `json:"date"`
}
