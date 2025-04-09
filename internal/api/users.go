package api

import (
	"encoding/json"
	"github.com/dabates/httpServer/internal/auth"
	"github.com/dabates/httpServer/internal/database"
	"github.com/dabates/httpServer/internal/types"
	"log"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request, config *types.ApiConfig) {
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
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("content-type", "application/json")
	w.Write(data)
}
