package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/sandbox"
)

func BenchmarkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
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

		result, err := sandbox.RunBenchmark(data)
		if err != nil {
			http.Error(w, "benchmark failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
