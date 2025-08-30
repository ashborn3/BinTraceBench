package api

import (
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/auth"
	"github.com/ashborn3/BinTraceBench/internal/database"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, db database.Database) {
	authMiddleware := auth.NewMiddleware(db)
	authHandler := auth.NewHandler(db)

	// Public routes
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, BinTraceBench!"))
	})
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	router.Get("/health", HealthHandler(db))
	router.Get("/ready", ReadinessHandler(db))

	// Authentication routes
	router.Post("/auth/register", authHandler.Register())
	router.Post("/auth/login", authHandler.Login())

	// Protected routes - require authentication
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		// Auth routes that require authentication
		r.Get("/auth/me", authHandler.Me())
		r.Post("/auth/logout", authHandler.Logout())

		// Binary analysis routes
		r.Post("/analyze", AnalyzeHandler(db))
		r.Get("/analyze", GetAnalysisResultsHandler(db))
		r.Get("/analyze/{id}", GetAnalysisResultHandler(db))
		r.Delete("/analyze/{id}", DeleteAnalysisResultHandler(db))

		// Benchmark routes
		r.Post("/bench", BenchmarkHandler(db))
		r.Get("/bench", GetBenchmarkResultsHandler(db))
		r.Get("/bench/{id}", GetBenchmarkResultHandler(db))
		r.Delete("/bench/{id}", DeleteBenchmarkResultHandler(db))

		// Process inspection routes - these can be public but are now protected
		r.Get("/proc/{pid}", InspectorHandler())
		r.Get("/proc/{pid}/files", OpenFileHandler())
		r.Get("/proc/{pid}/net", NetConnectionsHandler())
	})
}
