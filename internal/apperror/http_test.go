package apperror

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil error", nil, http.StatusOK},
		{"author not found", ErrAuthorNotFound, http.StatusNotFound},
		{"book not found", ErrBookNotFound, http.StatusNotFound},
		{"member not found", ErrMemberNotFound, http.StatusNotFound},
		{"user not found", ErrUserNotFound, http.StatusNotFound},
		{"loan not found", ErrLoanNotFound, http.StatusNotFound},
		{"author email exists", ErrAuthorEmailExists, http.StatusConflict},
		{"book isbn exists", ErrBookISBNExists, http.StatusConflict},
		{"member email exists", ErrMemberEmailExists, http.StatusConflict},
		{"book already on loan", ErrBookAlreadyOnLoan, http.StatusConflict},
		{"loan already returned", ErrLoanAlreadyReturned, http.StatusConflict},
		{"book title required", ErrBookTitleRequired, http.StatusBadRequest},
		{"invalid isbn", ErrInvalidISBN, http.StatusBadRequest},
		{"invalid name", ErrInvalidName, http.StatusBadRequest},
		{"invalid email", ErrInvalidEmail, http.StatusBadRequest},
		{"book author required", ErrBookAuthorRequired, http.StatusBadRequest},
		{"book id required", ErrBookIDRequired, http.StatusBadRequest},
		{"member id required", ErrMemberIDRequired, http.StatusBadRequest},
		{"invalid loan date", ErrInvalidLoanDate, http.StatusBadRequest},
		{"invalid due date", ErrInvalidDueDate, http.StatusBadRequest},
		{"invalid loan id", ErrInvalidLoanID, http.StatusBadRequest},
		{"unknown error", errors.New("unexpected db error"), http.StatusInternalServerError},
		{"member has max loans", ErrMemberHasMaxLoans, http.StatusConflict},
		{"invalid price", ErrInvalidPrice, http.StatusBadRequest},
		{"invalid genre", ErrInvalidGenre, http.StatusBadRequest},
		{"invalid published date", ErrInvalidPublishedDate, http.StatusBadRequest},
		{"invalid role", ErrInvalidRole, http.StatusBadRequest},
		{"invalid gender", ErrInvalidGender, http.StatusBadRequest},
		{"invalid password", ErrInvalidPassword, http.StatusBadRequest},
		{"email taken", ErrEmailTaken, http.StatusConflict},
		{"invalid credentials", ErrInvalidCredentials, http.StatusUnauthorized},
		{"unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"token expired", ErrTokenExpired, http.StatusUnauthorized},
		{"token invalid", ErrTokenInvalid, http.StatusUnauthorized},
		{"forbidden", ErrForbidden, http.StatusForbidden},
		{"context deadline exceeded", context.DeadlineExceeded, http.StatusGatewayTimeout},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := HTTPStatus(c.err)
			if got != c.want {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}
