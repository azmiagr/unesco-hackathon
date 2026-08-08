package model

import "github.com/google/uuid"

type GetUserParam struct {
	UserID   uuid.UUID `json:"-"`
	Email    string    `json:"-"`
	Username string    `json:"-"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegisterResponse struct {
	SessionToken string `json:"session_token"`
}
