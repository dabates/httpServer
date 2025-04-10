package api

import (
	"context"
	"encoding/json"
	"github.com/dabates/httpServer/internal/types"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type polkaRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserId string `json:"user_id"`
	} `json:"data"`
}

func PolkaWebhook(w http.ResponseWriter, r *http.Request, config *types.ApiConfig) {
	polkaData := polkaRequest{}
	err := json.NewDecoder(r.Body).Decode(&polkaData)
	if err != nil {
		log.Fatal(err)
	}

	if polkaData.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if polkaData.Event == "user.upgraded" {
		userId, err := uuid.Parse(polkaData.Data.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		_, err = config.Db.UpdateUserRedStatus(context.Background(), userId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}

	w.WriteHeader(http.StatusNotFound)
}
