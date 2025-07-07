package main

import (
	"fmt"
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	fmt.Println("Hello, BinTraceBench!")
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	api.RegisterRoutes(router)

	http.ListenAndServe(":8080", router)
}
