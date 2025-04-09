package api

import (
	"encoding/json"
	"fmt"
	"github.com/dabates/httpServer/internal/auth"
	"github.com/dabates/httpServer/internal/database"
	"github.com/dabates/httpServer/internal/types"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
)

func GetChirps(w http.ResponseWriter, r *http.Request, config *types.ApiConfig) {
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

		chirp, err := config.Db.GetChirp(r.Context(), id)
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

	chirps, err := config.Db.GetChirps(r.Context())
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

func Chirps(w http.ResponseWriter, r *http.Request, config *types.ApiConfig) {
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

	//Validate the jwt
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	userID, err := auth.ValidateJWT(token, config.Secret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	bodyData := reqBody{}
	err = json.NewDecoder(r.Body).Decode(&bodyData)
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

	chirp, err := config.Db.CreateChirp(r.Context(), database.CreateChirpParams{
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

func replaceBadWord(line string, badWord string) string {
	for _, word := range strings.Split(line, " ") {
		if strings.ToLower(word) == strings.ToLower(badWord) {
			line = strings.Replace(line, word, "****", -1)
		}
	}

	return line
}
