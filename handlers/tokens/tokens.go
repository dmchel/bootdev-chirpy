package tokens

import (
	"log"
	"net/http"
	"time"

	cfg "github.com/dmchel/bootdev-chirpy/config"
	"github.com/dmchel/bootdev-chirpy/internal/auth"
	"github.com/dmchel/bootdev-chirpy/utils"
	"github.com/google/uuid"
)

type TokensHandler struct {
	apiConfig *cfg.ApiConfig
}

func NewTokensHandler(apiConfig *cfg.ApiConfig) *TokensHandler {
	return &TokensHandler{apiConfig: apiConfig}
}

func (th *TokensHandler) RefreshAuthToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println("Failed to get refresh token from auth header", err)
		utils.UnauthorizedHandler(w, r)
		return
	}

	log.Println("Refresh token:", token)
	refreshToken, err := th.apiConfig.DBQueries.GetRefreshToken(r.Context(), token)
	if err != nil {
		log.Println("Couldn't read refresh token from DB", err)
		utils.UnauthorizedHandler(w, r)
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		log.Println("Refresh token expired")
		utils.UnauthorizedHandler(w, r)
		return
	}

	if refreshToken.RevokedAt.Valid && refreshToken.RevokedAt.Time.Before(time.Now()) {
		log.Println("Refresh token has been revoked")
		utils.UnauthorizedHandler(w, r)
		return
	}

	userId, err := uuid.Parse(refreshToken.UserID)
	if err != nil {
		log.Println("Invalid user ID returned from DB", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	authToken, err := auth.MakeJWT(userId, th.apiConfig.JWTSecret, 3600*time.Second)
	if err != nil {
		log.Println("Couldn't make JWT:", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	type tokenResponse struct {
		Token string `json:"token"`
	}
	var response tokenResponse
	response.Token = authToken

	log.Println("Refresh token for user", userId)
	utils.Respond(w, 200, response)
}

func (th *TokensHandler) RevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println("Failed to get refresh token from auth header", err)
		utils.UnauthorizedHandler(w, r)
		return
	}

	revokedToken, err := th.apiConfig.DBQueries.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		log.Println("Failed to revoke token")
		utils.InternalServerErrorHandler(w, r)
		return
	}
	log.Println("Token has been revoked:", revokedToken.Token, revokedToken.UserID)
	w.WriteHeader(204)
}
