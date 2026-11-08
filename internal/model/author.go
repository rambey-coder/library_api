package model

import "time"

type Author struct {
	ID        int64      `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Bio       *string    `json:"bio"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type CreateAuthorRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"  validate:"required"`
	Email     string `json:"email"      validate:"required,email"`
	Bio       string `json:"bio"        validate:"omitempty"`
}

type AuthorBookCount struct {
	ID        int64    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BookCount int    `json:"book_count"`
}