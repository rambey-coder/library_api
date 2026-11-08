package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeBookService struct {
	BookService
	getBookByIDFn func(ctx context.Context, id int64) (*model.Book, error)
}

func (f *fakeBookService) GetBookByID(ctx context.Context, id int64) (*model.Book, error) {
	return f.getBookByIDFn(ctx, id)
}

type fakeUserService struct {
	UserService
	createUserFn func(ctx context.Context, user model.CreateUserRequest) (*model.Users, error)
}

func (f *fakeUserService) CreateUser(ctx context.Context, user model.CreateUserRequest) (*model.Users, error) {
	return f.createUserFn(ctx, user)
}

func TestGetBookByID(t *testing.T) {
	// 1. Subtest for 200 OK
	t.Run("200 success", func(t *testing.T) {
		h := &Handler{
			book: &fakeBookService{
				getBookByIDFn: func(ctx context.Context, id int64) (*model.Book, error) {
					return &model.Book{ID: id, Title: "Clean Code"}, nil
				},
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/books/1", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()

		h.GetBook(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("want status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	// 2. Subtest for 404 Not Found + Error Envelope shape
	t.Run("404 not found and error envelope", func(t *testing.T) {
		h := &Handler{
			book: &fakeBookService{
				getBookByIDFn: func(ctx context.Context, id int64) (*model.Book, error) {
					return nil, apperror.ErrBookNotFound
				},
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/books/99", nil)
		req.SetPathValue("id", "99")
		rec := httptest.NewRecorder()

		h.GetBook(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("want status %d, got %d", http.StatusNotFound, rec.Code)
		}

		var resp utils.Response[any]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if resp.Success {
			t.Fatalf("want success=false, got true")
		}
		if resp.Error == nil {
			t.Fatalf("expected error object in response, got nil")
		}
		if resp.Error.Code != "Not Found" {
			t.Fatalf("want error code %q, got %q", "Not Found", resp.Error.Code)
		}
		if resp.Error.Message != apperror.ErrBookNotFound.Error() {
			t.Fatalf("want error message %q, got %q", apperror.ErrBookNotFound.Error(), resp.Error.Message)
		}
	})
}

func TestRegisterNeverContainsPasswordHash(t *testing.T) {
	secretHash := "$2a$10$secretBcryptHashValueThatMustNeverLeak"
	memberID := int64(1)

	h := &Handler{
		user: &fakeUserService{
			createUserFn: func(ctx context.Context, req model.CreateUserRequest) (*model.Users, error) {
				return &model.Users{
					ID:           1,
					FirstName:    req.FirstName,
					LastName:     req.LastName,
					Email:        req.Email,
					Role:         utils.RoleMember,
					MemberID:     &memberID,
					PasswordHash: secretHash,
				}, nil
			},
		},
	}

	body := `{"first_name":"Ada","last_name":"Lovelace","email":"ada@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("want status %d, got %d (body: %s)", http.StatusCreated, rec.Code, rec.Body.String())
	}

	rawBody := rec.Body.String()
	if strings.Contains(rawBody, "password_hash") || strings.Contains(rawBody, "password") {
		t.Fatalf("response JSON leaked password key: %s", rawBody)
	}
	if strings.Contains(rawBody, secretHash) {
		t.Fatalf("response JSON leaked password hash value: %s", rawBody)
	}
}
