package config

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/dmchel/bootdev-chirpy/internal/database"
	"github.com/dmchel/bootdev-chirpy/utils"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DBQueries      *database.Queries
	Platform       string
	JWTSecret      string
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)

	hits := cfg.FileserverHits.Load()
	html := fmt.Sprintf(`
<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
</html>`,
		hits)

	w.Write([]byte(html))
}

func (cfg *ApiConfig) ResetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		utils.ForbiddenHandler(w, r)
		return
	}

	result, err := cfg.DBQueries.DeleteAllUsers(r.Context())
	if err != nil {
		log.Println("Couldn't delete users", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	nrows, err := result.RowsAffected()
	if err != nil {
		log.Println("Couldn't delete users", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}
	log.Println("Users deleted: ", nrows)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.FileserverHits.Store(0)
	w.Write([]byte("Status: OK"))
}
