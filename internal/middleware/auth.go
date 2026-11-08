package middleware

import (
	"context"
	"errors"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/utils"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(w, r, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		tokenStr, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok {
			utils.ErrorResponse(w, r, http.StatusUnauthorized, "invalid token format")
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &model.Claims{}, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperror.ErrTokenInvalid
			}

			return []byte(os.Getenv("JWT_SECRET")), nil

		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				utils.ErrorHandler(w, r, apperror.ErrTokenExpired)
			} else {
				utils.ErrorHandler(w, r, apperror.ErrTokenInvalid)
			}

			return
		}

		if !token.Valid {
			utils.ErrorResponse(w, r, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(*model.Claims)
		if !ok {
			utils.ErrorResponse(w, r, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), utils.UserClaimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. read role from context (set by AuthMiddleware)
			// 2. if missing -> 401 (shouldn't happen if AuthMiddleware ran)
			claims, ok := r.Context().Value(utils.UserClaimsKey).(*model.Claims)
			if !ok || claims == nil {
				utils.ErrorHandler(w, r, apperror.ErrUnauthorized)
				return
			}

			// 3. if mismatch -> 403
			if !slices.Contains(roles, claims.Role) {
				utils.ErrorHandler(w, r, apperror.ErrForbidden)
				return
			}

			// 4. next
			next.ServeHTTP(w, r)
		})
	}
}

// Wraps a HandlerFunc with AuthMiddleware (and optionally RequireRole if roles are passed)
func Protect(fn http.HandlerFunc, roles ...string) http.Handler {
	var h http.Handler = fn

	if len(roles) > 0 {
		h = RequireRole(roles...)(h)
	}

	return AuthMiddleware(h)
}
