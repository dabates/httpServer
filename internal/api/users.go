package api

import (
	"context"
	"encoding/json"
	"github.com/dabates/httpServer/internal/auth"
	"github.com/dabates/httpServer/internal/database"
	"github.com/dabates/httpServer/internal/types"
	"log"
	"net/http"
)

type reqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type respBody struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
	ChirpyRed bool   `json:"is_chirpy_red"`
}

func CreateUser(w http.ResponseWriter, r *http.Request, config *types.ApiConfig) {

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

	user, err := config.Db.CreateUser(r.Context(), database.CreateUserParams{
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
		ChirpyRed: user.IsChirpyRed,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("content-type", "application/json")
	w.Write(data)
}

func UpdateUser(w http.ResponseWriter, r *http.Request, config *types.ApiConfig) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	userId, err := auth.ValidateJWT(token, config.Secret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	bodyData := reqBody{}
	err = json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	if bodyData.Email == "" {
		log.Fatal("Email is empty")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("No email was provided"))
		return
	}

	if bodyData.Password == "" {
		log.Fatal("Password is empty")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("No password was provided"))
		return
	}

	password, err := auth.HashPassword(bodyData.Password)
	if err != nil {
		log.Fatal(err)
	}

	user, err := config.Db.UpdateUser(context.Background(), database.UpdateUserParams{
		userId,
		bodyData.Email,
		password,
	})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	resp := respBody{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
		ChirpyRed: user.IsChirpyRed,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")
	w.Write(data)
}
