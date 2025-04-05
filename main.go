package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	apiConfig := apiConfig{}

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		w.WriteHeader(http.StatusOK)

		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type respBody struct {
			Error string `json:"error,omitempty"`
			Valid bool   `json:"valid"`
		}

		type reqBody struct {
			Body string `json:"Body"`
		}

		body := reqBody{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := respBody{
				Valid: false,
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
				Valid: false,
				Error: "Body is too long",
			}

			data, _ := json.Marshal(resp)
			w.Write(data)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("content-type", "application/json")

		resp := respBody{
			Valid: true,
			Error: "",
		}

		data, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := respBody{
				Valid: false,
				Error: err.Error(),
			}
			json.NewEncoder(w).Encode(resp)

			return
		}

		w.Write(data)
	})

	mux.HandleFunc("GET /admin/metrics", apiConfig.GetFileserverHits)
	mux.HandleFunc("POST /admin/reset", apiConfig.Reset)

	mux.Handle("/app/", apiConfig.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	log.Fatal(httpServer.ListenAndServe())
}
