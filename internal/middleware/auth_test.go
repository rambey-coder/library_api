package middleware

import (
	"context"
	"library_api/internal/model"
	"library_api/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func signTestToken(t *testing.T, secret string, claims model.Claims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return signed
}

func TestAuthMiddleware(t *testing.T) {
	const testSecret = "test-secret"
	t.Setenv("JWT_SECRET", testSecret)

	memberID := int64(42)
	validToken := signTestToken(t, testSecret, model.Claims{
		UserID:   7,
		MemberID: &memberID,
		Role:     utils.RoleMember,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-5 * time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})

	expiredToken := signTestToken(t, testSecret, model.Claims{
		UserID:   7,
		MemberID: &memberID,
		Role:     utils.RoleMember,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	})

	wrongSecretToken := signTestToken(t, "wrong-secret", model.Claims{
		UserID: 7,
		Role:   utils.RoleMember,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})

	tests := []struct {
		name           string
		authHeader     string
		wantStatus     int
		wantNextCalled bool
		wantUserID     int64
		wantRole       string
	}{
		{
			name:           "missing token returns 401",
			authHeader:     "",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "missing Bearer prefix returns 401",
			authHeader:     validToken,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "malformed token returns 401",
			authHeader:     "Bearer not.a.jwt",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "expired token returns 401",
			authHeader:     "Bearer " + expiredToken,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "wrong signature secret returns 401",
			authHeader:     "Bearer " + wrongSecretToken,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "valid token passes through and sets claims in context",
			authHeader:     "Bearer " + validToken,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantUserID:     7,
			wantRole:       utils.RoleMember,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			var gotClaims *model.Claims

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotClaims, _ = r.Context().Value(utils.UserClaimsKey).(*model.Claims)
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/books", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()

			AuthMiddleware(nextHandler).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("want status %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if nextCalled != tc.wantNextCalled {
				t.Fatalf("want nextCalled=%v, got %v", tc.wantNextCalled, nextCalled)
			}
			if tc.wantNextCalled {
				if gotClaims == nil {
					t.Fatal("expected non-nil *model.Claims in request context")
				}
				if gotClaims.UserID != tc.wantUserID {
					t.Errorf("want UserID %d, got %d", tc.wantUserID, gotClaims.UserID)
				}
				if gotClaims.Role != tc.wantRole {
					t.Errorf("want Role %q, got %q", tc.wantRole, gotClaims.Role)
				}
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		claims         *model.Claims
		allowedRoles   []string
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "missing claims in context returns 401",
			claims:         nil,
			allowedRoles:   []string{utils.RoleAdmin},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "wrong role returns 403",
			claims:         &model.Claims{UserID: 1, Role: utils.RoleMember},
			allowedRoles:   []string{utils.RoleAdmin, utils.RoleLibrarian},
			wantStatus:     http.StatusForbidden,
			wantNextCalled: false,
		},
		{
			name:           "right role passes through with 200",
			claims:         &model.Claims{UserID: 1, Role: utils.RoleAdmin},
			allowedRoles:   []string{utils.RoleAdmin, utils.RoleLibrarian},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/books", nil)
			if tc.claims != nil {
				ctx := context.WithValue(req.Context(), utils.UserClaimsKey, tc.claims)
				req = req.WithContext(ctx)
			}
			rec := httptest.NewRecorder()

			RequireRole(tc.allowedRoles...)(nextHandler).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("want status %d, got %d", tc.wantStatus, rec.Code)
			}
			if nextCalled != tc.wantNextCalled {
				t.Fatalf("want nextCalled=%v, got %v", tc.wantNextCalled, nextCalled)
			}
		})
	}
}
