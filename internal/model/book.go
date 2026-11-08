package model

import "time"

type Book struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	ISBN          string     `json:"isbn"`
	PublishedDate time.Time  `json:"published_date"`
	Genre         string     `json:"genre"`
	Price         int64      `json:"price"`
	CreatedAt     time.Time  `json:"created_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
	Authors       []*Author  `json:"authors"`
}

type CreateBookRequest struct {
	Title         string    `json:"title" validate:"required"`
	ISBN          string    `json:"isbn" validate:"required"`
	PublishedDate time.Time `json:"published_date" validate:"required"`
	Genre         string    `json:"genre" validate:"required"`
	Price         int64     `json:"price" validate:"required"`
	AuthorIds     []int64   `json:"author_ids" validate:"required"`
}

type UpdateBookRequest struct {
	Title         string    `json:"title" validate:"required"`
	ISBN          string    `json:"isbn" validate:"required"`
	PublishedDate time.Time `json:"published_date" validate:"required"`
	Genre         string    `json:"genre" validate:"required"`
	Price         int64     `json:"price" validate:"required"`
}
