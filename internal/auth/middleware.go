package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ashborn3/BinTraceBench/internal/database"
)

// contextKey is used for context keys to avoid collisions
type contextKey string

const (
	UserContextKey    contextKey = "user"
	SessionContextKey contextKey = "session"
)

type Middleware struct {
	db database.Database
}

func NewMiddleware(db database.Database) *Middleware {
	return &Middleware{
		db: db,
	}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		token := parts[1]
		if token == "" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Missing token")
			return
		}

		session, err := m.db.GetSessionByToken(token)
		if err != nil {
			m.writeErrorResponse(w, http.StatusInternalServerError, "Failed to validate token")
			return
		}

		if session == nil {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		user, err := m.db.GetUserByID(session.UserID)
		if err != nil {
			m.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user data")
			return
		}

		if user == nil {
			m.writeErrorResponse(w, http.StatusUnauthorized, "User not found")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, SessionContextKey, session)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// checks if user has required role
func (m *Middleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				m.writeErrorResponse(w, http.StatusUnauthorized, "User not found in context")
				return
			}

			if user.Role != role && role != "user" { // user is default, allow all authenticated users
				m.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				session, err := m.db.GetSessionByToken(token)
				if err == nil && session != nil {
					user, err := m.db.GetUserByID(session.UserID)
					if err == nil && user != nil {
						ctx := context.WithValue(r.Context(), UserContextKey, user)
						ctx = context.WithValue(ctx, SessionContextKey, session)
						r = r.WithContext(ctx)
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserFromContext(ctx context.Context) *database.User {
	if user, ok := ctx.Value(UserContextKey).(*database.User); ok {
		return user
	}
	return nil
}

func GetSessionFromContext(ctx context.Context) *database.Session {
	if session, ok := ctx.Value(SessionContextKey).(*database.Session); ok {
		return session
	}
	return nil
}

func (m *Middleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
