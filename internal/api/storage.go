package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ashborn3/BinTraceBench/internal/auth"
	"github.com/ashborn3/BinTraceBench/internal/database"
	"github.com/go-chi/chi/v5"
)

func GetAnalysisResultsHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		results, err := db.GetAnalysisResultsByUser(user.ID)
		if err != nil {
			http.Error(w, "Failed to get analysis results: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func GetAnalysisResultHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		result, err := db.GetAnalysisResult(id)
		if err != nil {
			http.Error(w, "Failed to get analysis result: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if result == nil {
			http.Error(w, "Analysis result not found", http.StatusNotFound)
			return
		}

		if result.UserID != user.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func DeleteAnalysisResultHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// First check if the result exists and belongs to the user
		result, err := db.GetAnalysisResult(id)
		if err != nil {
			http.Error(w, "Failed to get analysis result: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if result == nil {
			http.Error(w, "Analysis result not found", http.StatusNotFound)
			return
		}

		if result.UserID != user.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := db.DeleteAnalysisResult(id); err != nil {
			http.Error(w, "Failed to delete analysis result: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Analysis result deleted successfully"})
	}
}

func GetBenchmarkResultsHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		results, err := db.GetBenchmarkResultsByUser(user.ID)
		if err != nil {
			http.Error(w, "Failed to get benchmark results: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func GetBenchmarkResultHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		result, err := db.GetBenchmarkResult(id)
		if err != nil {
			http.Error(w, "Failed to get benchmark result: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if result == nil {
			http.Error(w, "Benchmark result not found", http.StatusNotFound)
			return
		}

		if result.UserID != user.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func DeleteBenchmarkResultHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		result, err := db.GetBenchmarkResult(id)
		if err != nil {
			http.Error(w, "Failed to get benchmark result: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if result == nil {
			http.Error(w, "Benchmark result not found", http.StatusNotFound)
			return
		}

		if result.UserID != user.ID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := db.DeleteBenchmarkResult(id); err != nil {
			http.Error(w, "Failed to delete benchmark result: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Benchmark result deleted successfully"})
	}
}
