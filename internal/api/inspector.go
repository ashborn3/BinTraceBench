package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/ashborn3/BinTraceBench/internal/inspector"
	"github.com/go-chi/chi/v5"
)

func InspectorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pidStr := chi.URLParam(r, "pid")
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			http.Error(w, "invalid pid", http.StatusBadRequest)
			return
		}
		if pid == -1 {
			pid = os.Getpid()
		}

		info, err := inspector.GetProcInfo(pid)
		if err != nil {
			http.Error(w, fmt.Sprintf("error inspecting process: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}
