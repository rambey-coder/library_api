package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Users struct {
	ID           int64      `json:"id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Email        string     `json:"email"`
	Role         string     `json:"role"`
	Phone        *string    `json:"phone"`
	Gender       *string    `json:"gender"`
	MemberID     *int64     `json:"member_id,omitempty"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type CreateUserRequest struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	Phone     *string `json:"phone"`
	Gender    *string `json:"gender"`
	Password  string  `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	Users
	Token string `json:"token"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	MemberID *int64 `json:"member_id,omitempty"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
