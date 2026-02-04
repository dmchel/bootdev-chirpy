package main

import (
	"log"
	"net/http"

	h "github.com/dmchel/bootdev-chirpy/handlers/healthcheck"
)

func main() {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/app", http.StripPrefix("/app", http.FileServer(http.Dir("./app"))))
	mux.HandleFunc("/healthz", h.HealthcheckHandler)

	log.Println("Starting server", server.Addr)
	log.Panic(server.ListenAndServe())
}
