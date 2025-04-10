package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dabates/httpServer/internal/auth"
	"github.com/dabates/httpServer/internal/database"
	"github.com/dabates/httpServer/internal/types"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sort"
	"strings"
)

type chirpsBody struct {
	Id        string `json:"id"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func GetChirps(w http.ResponseWriter, r *http.Request, config *types.ApiConfig) {
	id := r.PathValue("id")
	fmt.Println("ID:", id)

	sortDir := r.URL.Query().Get("sort")
	if sortDir == "" {
		sortDir = "asc"
	}

	fmt.Println("SORT DIR:", sortDir)

	author_id := r.URL.Query().Get("author_id")
	if len(author_id) > 0 {
		userId, err := uuid.Parse(author_id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		chirps, err := config.Db.GetChirpsByUser(context.Background(), userId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		sort.Slice(chirps, func(i, j int) bool {
			if sortDir == "asc" {
				return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
			}
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})

		w.WriteHeader(http.StatusOK)
		resp := make([]chirpsBody, len(chirps))
		for i, chirp := range chirps {
			resp[i] = chirpsBody{
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
		return
	}

	if id != "" {
		id, err := uuid.Parse(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		chirp, err := config.Db.GetChirp(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		resp := chirpsBody{
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

	sort.Slice(chirps, func(i, j int) bool {
		if sortDir == "asc" {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		}
		return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
	})

	resp := make([]chirpsBody, len(chirps))
	for i, chirp := range chirps {
		resp[i] = chirpsBody{
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
	resp := chirpsBody{
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

func DeleteChirp(w http.ResponseWriter, r *http.Request, config *types.ApiConfig) {
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
	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No id provided"))
		return
	}

	// verify the chirp is by this user
	chirp, err := config.Db.GetChirp(r.Context(), uuid.MustParse(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	if chirp.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Not allowed to delete this chirp"))

		return
	}

	err = config.Db.DeleteChirp(context.Background(), database.DeleteChirpParams{
		ID:     chirp.ID,
		UserID: userID,
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func replaceBadWord(line string, badWord string) string {
	for _, word := range strings.Split(line, " ") {
		if strings.ToLower(word) == strings.ToLower(badWord) {
			line = strings.Replace(line, word, "****", -1)
		}
	}

	return line
}
