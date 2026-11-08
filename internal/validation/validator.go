package validation

import (
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"regexp"
	"strings"
	"time"
)

var (
	emailRgx  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	isbnRegex = regexp.MustCompile(`^(?:\d{10}|\d{13})$`)
)

func ValidateCreateAuthorRequest(request model.CreateAuthorRequest) error {

	if strings.TrimSpace(request.FirstName) == "" || strings.TrimSpace(request.LastName) == "" {
		return apperror.ErrInvalidName
	}
	if !emailRgx.MatchString(request.Email) {
		return apperror.ErrInvalidEmail
	}

	return nil
}

func ValidateCreateBookRequest(request model.CreateBookRequest) error {
	if strings.TrimSpace(request.Title) == "" {
		return apperror.ErrBookTitleRequired
	}
	if !isbnRegex.MatchString(request.ISBN) {
		return apperror.ErrInvalidISBN
	}
	if len(request.AuthorIds) == 0 {
		return apperror.ErrBookAuthorRequired
	}
	if request.Price < 0 {
		return apperror.ErrInvalidPrice
	}
	if strings.TrimSpace(request.Genre) == "" {
		return apperror.ErrInvalidGenre
	}
	if request.PublishedDate.IsZero() {
		return apperror.ErrInvalidPublishedDate
	}

	return nil
}

func ValidateCreateMemberRequest(request model.CreateMemberRequest) error {

	if strings.TrimSpace(request.Email) == "" || !emailRgx.MatchString(request.Email) {
		return apperror.ErrInvalidEmail
	}
	if strings.TrimSpace(request.FirstName) == "" || strings.TrimSpace(request.LastName) == "" {
		return apperror.ErrInvalidName
	}
	return nil
}

func ValidateLoanRequest(request model.LoanReq, now time.Time) error {
	if request.BookID <= 0 {
		return apperror.ErrBookIDRequired
	}

	if request.MemberID <= 0 {
		return apperror.ErrMemberIDRequired
	}

	if request.LoanDate.IsZero() || request.LoanDate.Before(now) {
		return apperror.ErrInvalidLoanDate
	}

	if request.DueDate.IsZero() || !request.DueDate.After(request.LoanDate) {
		return apperror.ErrInvalidDueDate
	}
	if request.DueDate.After(request.LoanDate.AddDate(0, 3, 0)) {
		return apperror.ErrInvalidDueDate
	}

	return nil
}

func ValidateLoanReturn(request int64) error {
	if request <= 0 {
		return apperror.ErrInvalidLoanID
	}

	return nil
}

func IsValidRole(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case utils.RoleAdmin, utils.RoleMember, utils.RoleLibrarian:
		return true
	}
	return false
}

func IsValidGender(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case utils.Male, utils.Female:
		return true
	}
	return false
}

func ValidateCreateUserRequest(request model.CreateUserRequest) error {
	if strings.TrimSpace(request.FirstName) == "" || len(strings.TrimSpace(request.FirstName)) <= 2 {
		return apperror.ErrInvalidName
	}
	if strings.TrimSpace(request.LastName) == "" || len(strings.TrimSpace(request.LastName)) <= 2 {
		return apperror.ErrInvalidName
	}
	if strings.TrimSpace(request.Email) == "" || !emailRgx.MatchString(request.Email) {
		return apperror.ErrInvalidEmail
	}
	if strings.TrimSpace(request.Role) != "" {
		return apperror.ErrInvalidRole
	}
	if request.Gender != nil && !IsValidGender(*request.Gender) {
		return apperror.ErrInvalidGender
	}
	if strings.TrimSpace(request.Password) == "" || len(strings.TrimSpace(request.Password)) < 8 || len(strings.TrimSpace(request.Password)) > 70 {
		return apperror.ErrInvalidPassword
	}

	return nil
}

func ValidateLoginRequest(request model.LoginRequest) error {

	if strings.TrimSpace(request.Email) == "" || !emailRgx.MatchString(request.Email) {
		return apperror.ErrInvalidEmail
	}

	if strings.TrimSpace(request.Password) == "" || len(strings.TrimSpace(request.Password)) < 8 {
		return apperror.ErrInvalidPassword
	}

	return nil
}
