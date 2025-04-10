package main

import (
	"database/sql"
	"fmt"
	"github.com/dabates/httpServer/internal/api"
	"github.com/dabates/httpServer/internal/database"
	"github.com/dabates/httpServer/internal/types"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)
import _ "github.com/lib/pq"

func main() {
	apiConfig := types.ApiConfig{}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	platform := os.Getenv("PLATFORM")
	fmt.Println("\n\nplatform:", platform)
	apiConfig.Platform = platform

	secret := os.Getenv("SECRET")
	apiConfig.Secret = secret

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbQueries := database.New(db)
	apiConfig.Db = dbQueries

	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		w.WriteHeader(http.StatusOK)

		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		api.CreateUser(w, r, &apiConfig)
	})
	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		api.UpdateUser(w, r, &apiConfig)
	})

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		api.Chirps(w, r, &apiConfig)
	})
	mux.HandleFunc("GET /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		api.GetChirps(w, r, &apiConfig)
	})
	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		api.GetChirps(w, r, &apiConfig)
	})
	mux.HandleFunc("DELETE /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		api.DeleteChirp(w, r, &apiConfig)
	})

	// Webhooks
	mux.HandleFunc("POST /api/polka/webhooks", func(w http.ResponseWriter, r *http.Request) {
		api.PolkaWebhook(w, r, &apiConfig)
	})

	// auth
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		api.Login(w, r, &apiConfig)
	})
	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		api.Refresh(w, r, &apiConfig)
	})
	mux.HandleFunc("POST /api/revoke", func(w http.ResponseWriter, r *http.Request) {
		api.Revoke(w, r, &apiConfig)
	})

	mux.HandleFunc("GET /admin/metrics", apiConfig.GetFileserverHits)
	mux.HandleFunc("POST /admin/reset", apiConfig.Reset)

	mux.Handle("/app/", apiConfig.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	log.Fatal(httpServer.ListenAndServe())
}
