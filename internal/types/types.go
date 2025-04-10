package types

import (
	"context"
	"fmt"
	"github.com/dabates/httpServer/internal/database"
	"log"
	"net/http"
	"sync/atomic"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
	Platform       string
	Db             *database.Queries
	Secret         string
	PolkaApiKey    string
}

func (c *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (c *ApiConfig) GetFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, c.fileserverHits.Load())))
}

func (c *ApiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	if c.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	c.fileserverHits.Store(0)

	err := c.Db.DeleteUsers(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte("OK"))
}
