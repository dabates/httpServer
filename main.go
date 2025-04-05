package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
