package models

import (
	"time"
)

// User represents a user in the system.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Don't expose password hash in JSON responses
	FullName     string    `json:"full_name"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	Status       string    `json:"status"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserCreate struct {
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

type UserUpdate struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	FullName    string    `json:"full_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

type UserUpdatePass struct {
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type UserUpdateStatus struct {
	Email  string `json:"email"`
	Status string `json:"status"`
}

type GetAllUsers struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Status   string `json:"status"`
	Role     string `json:"role"`
}
