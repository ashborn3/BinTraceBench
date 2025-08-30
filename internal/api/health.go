package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ashborn3/BinTraceBench/internal/database"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
	Version   string    `json:"version"`
}

func HealthHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		dbStatus := "ok"
		if err := db.Ping(); err != nil {
			dbStatus = "error"
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		response := HealthResponse{
			Status:    dbStatus,
			Timestamp: time.Now(),
			Database:  dbStatus,
			Version:   "1.0.0",
		}

		json.NewEncoder(w).Encode(response)
	}
}

func ReadinessHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "Database not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	}
}
