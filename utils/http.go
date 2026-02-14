package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

const ContentTypeJson = "application/json; charset=utf-8"

func Respond(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", ContentTypeJson)
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("Respond failed", err)
	}
}

func RespondBytes(w http.ResponseWriter, code int, bytes []byte) {
	w.Header().Set("Content-Type", ContentTypeJson)
	w.WriteHeader(code)
	if _, err := w.Write(bytes); err != nil {
		log.Println("Respond failed", err)
	}
}

func ServiceUnavailableHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("503 Service Unavailable"))
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
}
