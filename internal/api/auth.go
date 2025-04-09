package api

import (
	"context"
	"encoding/json"
	"github.com/dabates/httpServer/internal/auth"
	"github.com/dabates/httpServer/internal/types"
	"log"
	"net/http"
	"time"
)

func Login(w http.ResponseWriter, r *http.Request, a *types.ApiConfig) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type respBody struct {
		Id           string `json:"id"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	bodyData := reqBody{}
	err := json.NewDecoder(r.Body).Decode(&bodyData)
	if err != nil {
		log.Fatal(err)
	}

	user, err := a.Db.GetUserByEmail(r.Context(), bodyData.Email)
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

	//Get token for auth
	token, err := auth.MakeJWT(user.ID, a.Secret, time.Duration(3600)*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	// and the refresh token
	refreshToken, err := auth.MakeRefreshToken(user.ID, a.Db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := respBody{
		Id:           user.ID.String(),
		CreatedAt:    user.CreatedAt.String(),
		UpdatedAt:    user.UpdatedAt.String(),
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("content-type", "application/json")
	w.Write(data)
}

func Refresh(w http.ResponseWriter, r *http.Request, a *types.ApiConfig) {
	type respBody struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	ok, _ := auth.ValidateRefreshToken(refreshToken, a.Db)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid refresh token"))
		return
	}

	user, err := a.Db.GetUserFromRefreshToken(context.Background(), refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	//Get token for auth
	token, err := auth.MakeJWT(user.UserID, a.Secret, time.Duration(3600)*time.Second)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := respBody{
		Token: token,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("content-type", "application/json")
	w.Write(data)
}

func Revoke(w http.ResponseWriter, r *http.Request, a *types.ApiConfig) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	err = a.Db.ExpireRefreshToken(context.Background(), refreshToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
