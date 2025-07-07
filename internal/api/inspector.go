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

func OpenFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pidStr := chi.URLParam(r, "pid")
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			http.Error(w, "invalid PID", http.StatusBadRequest)
			return
		}

		files, err := inspector.GetOpenFiles(pid)
		if err != nil {
			http.Error(w, "could not get open files: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	}
}

func NetConnectionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pidStr := chi.URLParam(r, "pid")
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			http.Error(w, "invalid PID", http.StatusBadRequest)
			return
		}

		conns, err := inspector.GetNetworkConnections(pid)
		if err != nil {
			http.Error(w, "could not read net connections: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conns)
	}
}
