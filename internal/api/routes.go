package api

import (
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/auth"
	"github.com/ashborn3/BinTraceBench/internal/database"
	"github.com/go-chi/chi/v5"
)

var api_docs string = `
Hello, BinTraceBench!

API Documentation:
  GET  /health        - Health check
  GET  /ready         - Readiness check
  POST /auth/register - Register new user
  POST /auth/login    - Login user
  GET  /auth/me       - Get current user info
  POST /auth/logout   - Logout user
  POST /analyze       - Analyze binary (with optional ?dynamic=true)
  GET  /analyze       - List all analysis results
  GET  /analyze/{id}  - Get specific analysis result
  POST /bench         - Benchmark binary (with optional ?trace=true)
  GET  /bench         - List all benchmark results
  GET  /bench/{id}    - Get specific benchmark result
  GET  /proc/{pid}    - Inspect process
  GET  /proc/{pid}/files - Get process open files
  GET  /proc/{pid}/net   - Get process network connections

All endpoints except /auth/register and /auth/login require authentication
Use Authorization: Bearer <token> header for authenticated requests
`

func RegisterRoutes(router chi.Router, db database.Database) {
	authMiddleware := auth.NewMiddleware(db)
	authHandler := auth.NewHandler(db)

	// Public routes
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(api_docs))
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
