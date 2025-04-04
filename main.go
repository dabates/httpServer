package main

import (
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

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		w.WriteHeader(http.StatusOK)

		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/metrics", apiConfig.GetFileserverHits)
	mux.HandleFunc("/reset", apiConfig.Reset)

	mux.Handle("/app/", apiConfig.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	log.Fatal(httpServer.ListenAndServe())
}
