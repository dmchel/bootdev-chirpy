package chirps

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	cfg "github.com/dmchel/bootdev-chirpy/config"
	"github.com/dmchel/bootdev-chirpy/internal/database"
	"github.com/dmchel/bootdev-chirpy/utils"
	"github.com/google/uuid"
)

type ChirpsHandler struct {
	apiConfig *cfg.ApiConfig
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func NewChirpsHandler(apiConfig *cfg.ApiConfig) *ChirpsHandler {
	return &ChirpsHandler{apiConfig: apiConfig}
}

func (h *ChirpsHandler) CreateChirp(w http.ResponseWriter, r *http.Request) {
	type createChirpRequest struct {
		Body string `json:"body"`
	}
	type errorResponse struct {
		Error string `json:"error"`
	}

	authUserId, err := utils.GetUserIdentityFromAuth(r, h.apiConfig.JWTSecret)
	if err != nil {
		log.Println("Auth failed", err)
		utils.UnauthorizedHandler(w, r)
		return
	}

	var chirpReq createChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&chirpReq); err != nil {
		log.Println("Failed to decode create chirp request", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	if len(chirpReq.Body) > 140 {
		var response errorResponse
		response.Error = "Chirp is too long"
		utils.Respond(w, http.StatusBadRequest, response)
		return
	}

	params := database.CreateChirpParams{
		Body:   cleanChirp(chirpReq.Body),
		UserID: authUserId.String(),
	}

	chirp, err := h.apiConfig.DBQueries.CreateChirp(r.Context(), params)
	if err != nil {
		log.Println("Failed to create new chirp", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	id, err := uuid.Parse(chirp.ID)
	if err != nil {
		log.Println("Invalid chirp ID returned from DB", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	userId, err := uuid.Parse(chirp.UserID)
	if err != nil {
		log.Println("Invalid user ID returned from DB", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	result := Chirp{
		ID:        id,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    userId,
	}

	log.Println("Chirp posted:", result)
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Println("Failed to marshal chirp", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	utils.RespondBytes(w, http.StatusCreated, bytes)
}

func (h *ChirpsHandler) GetChirps(w http.ResponseWriter, r *http.Request) {

	//authorId := r.URL.Query().Get("author_id")

	chirps, err := h.apiConfig.DBQueries.GetChirps(r.Context())
	if err != nil {
		log.Println("Failed to get chirps", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	var resultList []Chirp
	for _, chirp := range chirps {
		id, err := uuid.Parse(chirp.ID)
		if err != nil {
			log.Println("Invalid chirp ID returned from DB", err)
			utils.InternalServerErrorHandler(w, r)
			return
		}

		userId, err := uuid.Parse(chirp.UserID)
		if err != nil {
			log.Println("Invalid user ID returned from DB", err)
			utils.InternalServerErrorHandler(w, r)
			return
		}

		result := Chirp{
			ID:        id,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    userId,
		}
		resultList = append(resultList, result)
	}

	log.Println("Get all chirps:", resultList)
	bytes, err := json.Marshal(resultList)
	if err != nil {
		log.Println("Failed to marshal list of chirps", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	utils.RespondBytes(w, http.StatusOK, bytes)
}

func (h *ChirpsHandler) GetChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpId")
	chirp, err := h.apiConfig.DBQueries.GetChirp(r.Context(), chirpId)
	if err != nil {
		log.Println("Failed to get chirp", err)
		if err == sql.ErrNoRows {
			utils.NotFoundHandler(w, r)
		} else {
			utils.InternalServerErrorHandler(w, r)
		}
		return
	}

	id, err := uuid.Parse(chirp.ID)
	if err != nil {
		log.Println("Invalid chirp ID returned from DB", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	userId, err := uuid.Parse(chirp.UserID)
	if err != nil {
		log.Println("Invalid user ID returned from DB", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	result := Chirp{
		ID:        id,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    userId,
	}

	log.Println("Get chirp:", result)
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Println("Failed to marshal chirp", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	utils.RespondBytes(w, http.StatusOK, bytes)
}

func (h *ChirpsHandler) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpId")
	authUserId, err := utils.GetUserIdentityFromAuth(r, h.apiConfig.JWTSecret)
	if err != nil {
		log.Println("Auth failed", err)
		utils.UnauthorizedHandler(w, r)
		return
	}

	chirp, err := h.apiConfig.DBQueries.GetChirp(r.Context(), chirpId)
	if err != nil {
		log.Println("Failed to get chirp", err)
		if err == sql.ErrNoRows {
			utils.NotFoundHandler(w, r)
		} else {
			utils.InternalServerErrorHandler(w, r)
		}
		return
	}

	if chirp.UserID != authUserId.String() {
		log.Printf("User %s tried to delete someone else's chirp!", authUserId.String())
		utils.ForbiddenHandler(w, r)
		return
	}

	chirp, err = h.apiConfig.DBQueries.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		utils.InternalServerErrorHandler(w, r)
		return
	}

	log.Println("Chirp has been deleted: ", chirp)
	utils.Respond(w, http.StatusNoContent, nil)
}

func cleanChirp(body string) string {
	if len(body) == 0 {
		return ""
	}

	nastyWords := [3]string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(body, " ")
	for i, word := range words {
		lw := strings.ToLower(word)
		for _, nword := range nastyWords {
			if lw == nword {
				words[i] = "****"
				break
			}
		}
	}

	return strings.Join(words, " ")
}
