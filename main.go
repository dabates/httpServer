package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dabates/httpServer/internal/auth"
	"github.com/dabates/httpServer/internal/database"
	"github.com/google/uuid"
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
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		users(w, r, &apiConfig)
	})
	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirps(w, r, &apiConfig)
	})
	mux.HandleFunc("GET /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		getChirps(w, r, &apiConfig)
	})
	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		getChirps(w, r, &apiConfig)
	})
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		login(w, r, &apiConfig)
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

func getChirps(w http.ResponseWriter, r *http.Request, config *apiConfig) {
	type respBody struct {
		Id        string `json:"id"`
		Body      string `json:"body"`
		UserId    string `json:"user_id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	id := r.PathValue("id")
	fmt.Println("ID:", id)

	if id != "" {
		id, err := uuid.Parse(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		chirp, err := config.db.GetChirp(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		resp := respBody{
			Id:        chirp.ID.String(),
			Body:      chirp.Body,
			UserId:    chirp.UserID.String(),
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
		}
		data, err := json.Marshal(resp)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("content-type", "application/json")
		w.Write(data)
		return
	}

	chirps, err := config.db.GetChirps(r.Context())
	if err != nil {
		log.Fatal(err)
	}
	resp := make([]respBody, len(chirps))
	for i, chirp := range chirps {
		resp[i] = respBody{
			Id:        chirp.ID.String(),
			Body:      chirp.Body,
			UserId:    chirp.UserID.String(),
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
		}
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("content-type", "application/json")
	w.Write(data)

}

func chirps(w http.ResponseWriter, r *http.Request, config *apiConfig) {
	type reqBody struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}

	type respBody struct {
		Id        string `json:"id"`
		Body      string `json:"body"`
		UserId    string `json:"user_id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	bodyData := reqBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
	}

	if len(bodyData.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal("Body is too long")
	}

	line := bodyData.Body
	line = replaceBadWord(line, "kerfuffle")
	line = replaceBadWord(line, "sharbert")
	line = replaceBadWord(line, "fornax")

	userID, err := uuid.Parse(bodyData.UserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
	}

	chirp, err := config.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   line,
		UserID: userID,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("content-type", "application/json")
	resp := respBody{
		Id:        chirp.ID.String(),
		Body:      chirp.Body,
		UserId:    chirp.UserID.String(),
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	bodyData := reqBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		log.Fatal(err)
	}

	if bodyData.Email == "" {
		log.Fatal("Email is empty")
	}

	password, err := auth.HashPassword(bodyData.Password)
	if err != nil {
		log.Fatal(err)
	}

	user, err := config.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          bodyData.Email,
		HashedPassword: password,
	})

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

func login(w http.ResponseWriter, r *http.Request, a *apiConfig) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type respBody struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	bodyData := reqBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		log.Fatal(err)
	}

	user, err := a.db.GetUserByEmail(r.Context(), bodyData.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	ok := auth.CheckPasswordHash(bodyData.Password, user.HashedPassword)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid password"))
		return
	}

	w.WriteHeader(http.StatusOK)
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

	w.Header().Set("content-type", "application/json")
	w.Write(data)
}
