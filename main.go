package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dabates/httpServer/internal/database"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"
)
import _ "github.com/lib/pq"

func main() {
	apiConfig := apiConfig{}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	platform := os.Getenv("PLATFORM")
	fmt.Println("\n\nplatform:", platform)
	apiConfig.platform = platform

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbQueries := database.New(db)
	apiConfig.db = dbQueries

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
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("In post for users")
		users(w, r, &apiConfig)
	})

	mux.HandleFunc("GET /admin/metrics", apiConfig.GetFileserverHits)
	mux.HandleFunc("POST /admin/reset", apiConfig.Reset)

	mux.Handle("/app/", apiConfig.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	log.Fatal(httpServer.ListenAndServe())
}

func replaceBadWord(line string, badWord string) string {
	for _, word := range strings.Split(line, " ") {
		if strings.ToLower(word) == strings.ToLower(badWord) {
			line = strings.Replace(line, word, "****", -1)
		}
	}

	return line
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type respBody struct {
		Error       string `json:"error,omitempty"`
		CleanedBody string `json:"cleaned_body,omitempty"`
	}

	type reqBody struct {
		Body string `json:"Body"`
	}

	body := reqBody{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := respBody{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(resp)

		return
	}
	fmt.Println("got:")
	fmt.Println(body.Body)

	if len(body.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		resp := respBody{
			Error: "Body is too long",
		}

		data, _ := json.Marshal(resp)
		w.Write(data)
		return
	}

	line := body.Body
	line = replaceBadWord(line, "kerfuffle")
	line = replaceBadWord(line, "sharbert")
	line = replaceBadWord(line, "fornax")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")

	resp := respBody{
		CleanedBody: line,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := respBody{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(resp)

		return
	}

	w.Write(data)
}

func users(w http.ResponseWriter, r *http.Request, config *apiConfig) {
	type respBody struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	type reqBody struct {
		Email string `json:"email"`
	}

	bodyData := reqBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		log.Fatal(err)
	}

	if bodyData.Email == "" {
		log.Fatal("Email is empty")
	}

	user, err := config.db.CreateUser(r.Context(), bodyData.Email)
	if err != nil {
		log.Fatal(err)
	}

	resp := respBody{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("content-type", "application/json")
	w.Write(data)
}
