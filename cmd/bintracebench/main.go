package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/api"
	"github.com/ashborn3/BinTraceBench/internal/config"
	"github.com/ashborn3/BinTraceBench/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	fmt.Println("Starting BinTraceBench Server...")

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	fmt.Printf("Database: %s\n", cfg.Database.Type)
	if cfg.Database.Type == "sqlite" {
		fmt.Printf("SQLite Path: %s\n", cfg.Database.SQLite.Path)
	} else {
		fmt.Printf("PostgreSQL: %s:%d/%s\n",
			cfg.Database.Postgres.Host,
			cfg.Database.Postgres.Port,
			cfg.Database.Postgres.DBName)
	}

	dbFactory := database.NewFactory()
	db, err := dbFactory.Create(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("Database connected successfully")

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	// Add CORS middleware for web frontend
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	api.RegisterRoutes(router, db)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server starting on http://%s\n", addr)
	fmt.Println("API Documentation:")
	fmt.Println("  POST /auth/register - Register new user")
	fmt.Println("  POST /auth/login    - Login user")
	fmt.Println("  GET  /auth/me       - Get current user info")
	fmt.Println("  POST /auth/logout   - Logout user")
	fmt.Println("  POST /analyze       - Analyze binary (with optional ?dynamic=true)")
	fmt.Println("  GET  /analyze       - List all analysis results")
	fmt.Println("  GET  /analyze/{id}  - Get specific analysis result")
	fmt.Println("  POST /bench         - Benchmark binary (with optional ?trace=true)")
	fmt.Println("  GET  /bench         - List all benchmark results")
	fmt.Println("  GET  /bench/{id}    - Get specific benchmark result")
	fmt.Println("  GET  /proc/{pid}    - Inspect process")
	fmt.Println("  GET  /proc/{pid}/files - Get process open files")
	fmt.Println("  GET  /proc/{pid}/net   - Get process network connections")
	fmt.Println()
	fmt.Println("All endpoints except /auth/register and /auth/login require authentication")
	fmt.Println("Use Authorization: Bearer <token> header for authenticated requests")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
