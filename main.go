package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	cfg "github.com/dmchel/bootdev-chirpy/config"
	"github.com/dmchel/bootdev-chirpy/handlers/chirps"
	h "github.com/dmchel/bootdev-chirpy/handlers/healthcheck"
	"github.com/dmchel/bootdev-chirpy/handlers/users"
	"github.com/dmchel/bootdev-chirpy/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Panic("Failed to open DB connection.", err)
	}

	dbQueries := database.New(db)

	apiCfg := cfg.ApiConfig{DBQueries: dbQueries, Platform: platform}
	users := users.NewUserHandler(&apiCfg)
	chirps := chirps.NewChirpsHandler(&apiCfg)
	fs := http.FileServer(http.Dir("./app"))

	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.MiddlewareMetricsInc(fs)))
	mux.HandleFunc("GET /api/healthz", h.HealthcheckHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetMetricsHandler)
	mux.HandleFunc("POST /api/users", users.CreateUser)
	mux.HandleFunc("POST /api/chirps", chirps.CreateChirp)
	mux.HandleFunc("GET /api/chirps", chirps.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", chirps.GetChirp)

	log.Println("Starting server", server.Addr)
	log.Panic(server.ListenAndServe())
}
