package validation_test

import (
	"errors"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/validation"
	"testing"
	"time"
)

func TestValidateCreateAuthor(t *testing.T) {
	cases := []struct {
		name string
		req  model.CreateAuthorRequest
		want error
	}{
		{"empty first name", model.CreateAuthorRequest{FirstName: " ", LastName: "Ade", Email: "rambey@mail.co"}, apperror.ErrInvalidName},
		{"empty last name", model.CreateAuthorRequest{FirstName: "Ola", LastName: "", Email: "rambey@mail.co"}, apperror.ErrInvalidName},
		{"bad email", model.CreateAuthorRequest{FirstName: "Ola", LastName: "Ade", Email: "rambey"}, apperror.ErrInvalidEmail},
		{"valid", model.CreateAuthorRequest{FirstName: "Ramadan", LastName: "Bello", Email: "rambeybello@gmail.com"}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := validation.ValidateCreateAuthorRequest(c.req)
			if !errors.Is(got, c.want) && !(got == nil && c.want == nil) {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}

func TestValidateCreateMember(t *testing.T) {
	cases := []struct {
		name string
		req  model.CreateMemberRequest
		want error
	}{
		{"empty email", model.CreateMemberRequest{FirstName: "Ola", LastName: "Ade", Email: ""}, apperror.ErrInvalidEmail},
		{"bad email", model.CreateMemberRequest{FirstName: "Ola", LastName: "Ade", Email: "invalid-email"}, apperror.ErrInvalidEmail},
		{"empty first name", model.CreateMemberRequest{FirstName: " ", LastName: "Ade", Email: "ola@mail.co"}, apperror.ErrInvalidName},
		{"empty last name", model.CreateMemberRequest{FirstName: "Ola", LastName: " ", Email: "ola@mail.co"}, apperror.ErrInvalidName},
		{"valid", model.CreateMemberRequest{FirstName: "Ola", LastName: "Ade", Email: "ola@mail.co"}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := validation.ValidateCreateMemberRequest(c.req)
			if !errors.Is(got, c.want) && !(got == nil && c.want == nil) {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}

func TestValidateLoanRequest(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	loanDate := now.Add(24 * time.Hour)
	validDueDate := loanDate.AddDate(0, 1, 0)

	cases := []struct {
		name string
		req  model.LoanReq
		want error
	}{
		{"missing book id", model.LoanReq{BookID: 0, MemberID: 1, LoanDate: loanDate, DueDate: validDueDate}, apperror.ErrBookIDRequired},
		{"missing member id", model.LoanReq{BookID: 1, MemberID: 0, LoanDate: loanDate, DueDate: validDueDate}, apperror.ErrMemberIDRequired},
		{"past loan date", model.LoanReq{BookID: 1, MemberID: 1, LoanDate: now.Add(-24 * time.Hour), DueDate: validDueDate}, apperror.ErrInvalidLoanDate},
		{"due date before loan date", model.LoanReq{BookID: 1, MemberID: 1, LoanDate: loanDate, DueDate: loanDate.Add(-24 * time.Hour)}, apperror.ErrInvalidDueDate},
		{"due date more than 3 months", model.LoanReq{BookID: 1, MemberID: 1, LoanDate: loanDate, DueDate: loanDate.AddDate(0, 4, 0)}, apperror.ErrInvalidDueDate},
		{"valid", model.LoanReq{BookID: 1, MemberID: 1, LoanDate: loanDate, DueDate: validDueDate}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := validation.ValidateLoanRequest(c.req, now)
			if !errors.Is(got, c.want) && !(got == nil && c.want == nil) {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}

func TestValidateLoanReturn(t *testing.T) {
	cases := []struct {
		name   string
		loanID int64
		want   error
	}{
		{"zero loan id", 0, apperror.ErrInvalidLoanID},
		{"negative loan id", -1, apperror.ErrInvalidLoanID},
		{"valid loan id", 1, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := validation.ValidateLoanReturn(c.loanID)
			if !errors.Is(got, c.want) && !(got == nil && c.want == nil) {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}

func TestValidateCreateBookRequest(t *testing.T) {
	cases := []struct {
		name string
		req  model.CreateBookRequest
		want error
	}{
		{"empty title", model.CreateBookRequest{Title: " ", ISBN: "1234567890", PublishedDate: time.Now(), AuthorIds: []int64{1}, Genre: "Fiction", Price: 10}, apperror.ErrBookTitleRequired},
		{"empty isbn", model.CreateBookRequest{Title: "Book", ISBN: "", PublishedDate: time.Now(), AuthorIds: []int64{1}, Genre: "Fiction", Price: 10}, apperror.ErrInvalidISBN},
		{"invalid isbn", model.CreateBookRequest{Title: "Book", ISBN: "invalid-isbn", PublishedDate: time.Now(), AuthorIds: []int64{1}, Genre: "Fiction", Price: 10}, apperror.ErrInvalidISBN},
		{"missing authors", model.CreateBookRequest{Title: "Book", ISBN: "1234567890", PublishedDate: time.Now(), AuthorIds: []int64{}, Genre: "Fiction", Price: 10}, apperror.ErrBookAuthorRequired},
		{"negative price", model.CreateBookRequest{Title: "Book", ISBN: "1234567890", PublishedDate: time.Now(), AuthorIds: []int64{1}, Genre: "Fiction", Price: -5}, apperror.ErrInvalidPrice},
		{"empty genre", model.CreateBookRequest{Title: "Book", ISBN: "1234567890", PublishedDate: time.Now(), AuthorIds: []int64{1}, Genre: " ", Price: 10}, apperror.ErrInvalidGenre},
		{"zero published date", model.CreateBookRequest{Title: "Book", ISBN: "1234567890", PublishedDate: time.Time{}, AuthorIds: []int64{1}, Genre: "Fiction", Price: 10}, apperror.ErrInvalidPublishedDate},
		{"valid", model.CreateBookRequest{Title: "Book", ISBN: "1234567890", PublishedDate: time.Now(), AuthorIds: []int64{1}, Genre: "Fiction", Price: 10}, nil},
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
