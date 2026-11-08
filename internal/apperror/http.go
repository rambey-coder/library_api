package apperror

import (
	"context"
	"errors"
	"net/http"
)

func HTTPStatus(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.Is(err, ErrAuthorNotFound),
		errors.Is(err, ErrBookNotFound),
		errors.Is(err, ErrMemberNotFound),
		errors.Is(err, ErrUserNotFound),
		errors.Is(err, ErrLoanNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrAuthorEmailExists),
		errors.Is(err, ErrBookISBNExists),
		errors.Is(err, ErrMemberEmailExists),
		errors.Is(err, ErrBookAlreadyOnLoan),
		errors.Is(err, ErrLoanAlreadyReturned),
		errors.Is(err, ErrEmailTaken),
		errors.Is(err, ErrMemberHasMaxLoans):
		return http.StatusConflict
	case errors.Is(err, ErrBookTitleRequired),
		errors.Is(err, ErrInvalidISBN),
		errors.Is(err, ErrInvalidName),
		errors.Is(err, ErrInvalidEmail),
		errors.Is(err, ErrBookAuthorRequired),
		errors.Is(err, ErrBookIDRequired),
		errors.Is(err, ErrMemberIDRequired),
		errors.Is(err, ErrInvalidLoanDate),
		errors.Is(err, ErrInvalidDueDate),
		errors.Is(err, ErrInvalidPrice),
		errors.Is(err, ErrInvalidGenre),
		errors.Is(err, ErrInvalidPublishedDate),
		errors.Is(err, ErrInvalidRole),
		errors.Is(err, ErrInvalidGender),
		errors.Is(err, ErrInvalidPassword),
		errors.Is(err, ErrInvalidLoanID):
		return http.StatusBadRequest
	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrUnauthorized),
		errors.Is(err, ErrTokenExpired),
		errors.Is(err, ErrTokenInvalid):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
