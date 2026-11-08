package service

import (
	"errors"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/validation"
	"testing"
	"time"
)

func TestValidateCreateBook(t *testing.T) {
	cases := []struct {
		name string
		req  model.CreateBookRequest
		want error
	}{
		{"empty title", model.CreateBookRequest{Title: ""}, apperror.ErrBookTitleRequired},
		{"bad isbn", model.CreateBookRequest{Title: "x", ISBN: "abc"}, apperror.ErrInvalidISBN},
		{"no authors", model.CreateBookRequest{Title: "x", ISBN: "9780000000001"}, apperror.ErrBookAuthorRequired},
		{"negative price", model.CreateBookRequest{Title: "x", ISBN: "9780000000001", AuthorIds: []int64{1}, Price: -1}, apperror.ErrInvalidPrice},
		{"empty genre", model.CreateBookRequest{Title: "x", ISBN: "9780000000001", AuthorIds: []int64{1}, Price: 1, Genre: " "}, apperror.ErrInvalidGenre},
		{"zero published date", model.CreateBookRequest{Title: "x", ISBN: "9780000000001", AuthorIds: []int64{1}, Price: 1, Genre: "Fiction"}, apperror.ErrInvalidPublishedDate},
		{"valid", model.CreateBookRequest{Title: "x", ISBN: "9780000000001", AuthorIds: []int64{1}, Price: 1, Genre: "Fiction", PublishedDate: time.Now()}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := validation.ValidateCreateBookRequest(c.req)
			if !errors.Is(got, c.want) && !(got == nil && c.want == nil) {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}
