package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (c *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) GetFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(fmt.Sprintf("Hits: %d\n", c.fileserverHits.Load())))
}

func (c *apiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	c.fileserverHits.Store(0)

	w.Write([]byte("OK"))
}
