package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserWithPassword struct {
	User         User
	PasswordHash string
}

type CreateUserInput struct {
	Email        string
	PasswordHash string
	FullName     string
	Role         string
}

type UpdateUserRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	FullName *string `json:"full_name"`
}
