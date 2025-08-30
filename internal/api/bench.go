package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/auth"
	"github.com/ashborn3/BinTraceBench/internal/database"
	"github.com/ashborn3/BinTraceBench/internal/sandbox"
)

type BenchmarkResponse struct {
	ID     int                  `json:"id,omitempty"`
	Result *sandbox.BenchResult `json:"result"`
	Cached bool                 `json:"cached,omitempty"`
}

func BenchmarkHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "could not read file", http.StatusInternalServerError)
			return
		}

		trace := r.URL.Query().Get("trace") == "true"
		fileHash := auth.GenerateFileHash(data)
		filename := header.Filename

		cached, err := db.GetBenchmarkResultByHash(user.ID, fileHash)
		if err == nil && cached != nil && cached.WithTrace == trace {
			// Return cached result if trace requirement matches
			response := BenchmarkResponse{
				ID:     cached.ID,
				Result: cached.Result,
				Cached: true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		var result *sandbox.BenchResult
		if trace {
			result, err = sandbox.RunBenchmarkWithTrace(data)
		} else {
			result, err = sandbox.RunBenchmark(data) // very unsafe, minimal hardening, why even do this?
		}
		if err != nil {
			http.Error(w, "benchmark failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		benchmarkResult := &database.BenchmarkResult{
			UserID:    user.ID,
			Filename:  filename,
			FileHash:  fileHash,
			Result:    result,
			WithTrace: trace,
		}

		if err := db.SaveBenchmarkResult(benchmarkResult); err != nil {
			// Log error but don't fail the request
		}

		w.Header().Set("Content-Type", "application/json")
		response := BenchmarkResponse{
			ID:     benchmarkResult.ID,
			Result: result,
			Cached: false,
		}
		json.NewEncoder(w).Encode(response)
	}
}
