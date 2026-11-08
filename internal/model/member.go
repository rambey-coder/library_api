package model

import "time"

type Member struct {
	ID            int64      `json:"id"`
	UserID        *int64     `json:"user_id,omitempty"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone"`
	CreatedAt     time.Time  `json:"created_at"`
	RemovedAt     *time.Time `json:"removed_at"`
	OpenLoanCount int        `json:"open_loan_count"`
}

type CreateMemberRequest struct {
	FirstName string  `json:"first_name" validate:"required"`
	LastName  string  `json:"last_name" validate:"required"`
	Email     string  `json:"email" validate:"required,email"`
	Phone     *string `json:"phone" validate:"omitempty"`
}
