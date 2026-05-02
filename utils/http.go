package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dmchel/bootdev-chirpy/internal/auth"
	"github.com/google/uuid"
)

const ContentTypeJson = "application/json; charset=utf-8"

func Respond(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", ContentTypeJson)
	w.WriteHeader(code)
	if data == nil {
		return
	}
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

func ForbiddenHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("403 Forbidden"))
}

func UnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("401 Unauthorized"))
}

func BadRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("400 Bad Request"))
}

func GetUserIdentityFromAuth(r *http.Request, tokenSecret string) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.UUID{}, err
	}

	return auth.ValidateJWT(token, tokenSecret)
}
