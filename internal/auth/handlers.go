package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ashborn3/BinTraceBench/internal/database"
	"github.com/ashborn3/BinTraceBench/pkg/logging"
)

type Handler struct {
	db      database.Database
	service *Service
}

func NewHandler(db database.Database) *Handler {
	return &Handler{
		db:      db,
		service: &Service{},
	}
}

func (h *Handler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate input
		if req.Username == "" || req.Password == "" || req.Email == "" {
			h.writeErrorResponse(w, http.StatusBadRequest, "Username, password, and email are required")
			return
		}

		if len(req.Username) < 3 || len(req.Username) > 50 {
			h.writeErrorResponse(w, http.StatusBadRequest, "Username must be between 3 and 50 characters")
			return
		}

		if len(req.Password) < 6 {
			h.writeErrorResponse(w, http.StatusBadRequest, "Password must be at least 6 characters")
			return
		}

		existingUser, err := h.db.GetUserByUsername(req.Username)
		if err != nil {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to check existing user")
			return
		}

		if existingUser != nil {
			h.writeErrorResponse(w, http.StatusConflict, "Username already exists")
			return
		}

		hashedPassword, err := h.service.HashPassword(req.Password)
		if err != nil {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		user := &database.User{
			Username: req.Username,
			Password: hashedPassword,
			Email:    req.Email,
			Role:     "user",
		}

		if err := h.db.CreateUser(user); err != nil {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Return success (don't include password)
		response := User{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			Created:  user.Created.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Username == "" || req.Password == "" {
			h.writeErrorResponse(w, http.StatusBadRequest, "Username and password are required")
			return
		}

		user, err := h.db.GetUserByUsername(req.Username)
		if err != nil {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
			return
		}

		if user == nil {
			logging.Warn("Login attempt with invalid username", "username", req.Username, "ip", r.RemoteAddr)
			h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		if !h.service.CheckPassword(req.Password, user.Password) {
			logging.Warn("Login attempt with invalid password", "username", req.Username, "ip", r.RemoteAddr)
			h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		token, err := h.service.GenerateToken()
		if err != nil {
			logging.Error("Failed to generate token", "error", err, "username", req.Username)
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		logging.Info("User logged in", "username", req.Username, "ip", r.RemoteAddr)

		expires := SessionExpiry()
		session := &database.Session{
			UserID:  user.ID,
			Token:   token,
			Expires: expires,
		}

		if err := h.db.CreateSession(session); err != nil {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create session")
			return
		}

		response := LoginResponse{
			Token:   token,
			Expires: expires.Format(time.RFC3339),
			User: User{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
				Role:     user.Role,
				Created:  user.Created.Format(time.RFC3339),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := GetSessionFromContext(r.Context())
		if session == nil {
			h.writeErrorResponse(w, http.StatusUnauthorized, "No active session")
			return
		}

		if err := h.db.DeleteSession(session.Token); err != nil {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to logout")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
	}
}

// Me returns current user info
func (h *Handler) Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			h.writeErrorResponse(w, http.StatusUnauthorized, "User not found")
			return
		}

		response := User{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			Created:  user.Created.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
