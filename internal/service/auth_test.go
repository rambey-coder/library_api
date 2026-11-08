package service

import (
	"context"
	"errors"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct {
	UserRepository
	getUserByEmailFn func(ctx context.Context, email string) (*model.Users, error)
}

func (f *fakeUserRepo) GetUserByEmail(ctx context.Context, email string) (*model.Users, error) {
	return f.getUserByEmailFn(ctx, email)
}

func TestLogin(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	correctPassword := "password123"
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	existingUser := &model.Users{
		ID:           1,
		FirstName:    "Ada",
		LastName:     "Lovelace",
		Email:        "ada@example.com",
		Role:         utils.RoleMember,
		PasswordHash: string(hashedBytes),
	}

	tests := []struct {
		name      string
		req       model.LoginRequest
		repoFn    func(ctx context.Context, email string) (*model.Users, error)
		wantErr   error
		wantToken bool
	}{
		{
			name: "right password returns token and user",
			req: model.LoginRequest{
				Email:    "ada@example.com",
				Password: correctPassword,
			},
			repoFn: func(ctx context.Context, email string) (*model.Users, error) {
				return existingUser, nil
			},
			wantErr:   nil,
			wantToken: true,
		},
		{
			name: "wrong password returns ErrInvalidCredentials",
			req: model.LoginRequest{
				Email:    "ada@example.com",
				Password: "wrongpassword123",
			},
			repoFn: func(ctx context.Context, email string) (*model.Users, error) {
				return existingUser, nil
			},
			wantErr:   apperror.ErrInvalidCredentials,
			wantToken: false,
		},
		{
			name: "nonexistent user returns ErrInvalidCredentials",
			req: model.LoginRequest{
				Email:    "nobody@example.com",
				Password: correctPassword,
			},
			repoFn: func(ctx context.Context, email string) (*model.Users, error) {
				return nil, apperror.ErrInvalidCredentials
			},
			wantErr:   apperror.ErrInvalidCredentials,
			wantToken: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewUserService(&fakeUserRepo{
				getUserByEmailFn: tc.repoFn,
			})

			res, err := svc.Login(context.Background(), tc.req)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want error %v, got %v", tc.wantErr, err)
			}

			if tc.wantToken {
				if res == nil || res.Token == "" {
					t.Fatal("expected non-empty JWT token on successful login")
				}
				if res.Email != existingUser.Email {
					t.Errorf("want email %q, got %q", existingUser.Email, res.Email)
				}
			}
		})
	}
}

