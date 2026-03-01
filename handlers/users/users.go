package users

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	cfg "github.com/dmchel/bootdev-chirpy/config"
	"github.com/dmchel/bootdev-chirpy/utils"
	"github.com/google/uuid"
)

type UsersHandler struct {
	apiConfig *cfg.ApiConfig
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func NewUserHandler(apiConfig *cfg.ApiConfig) *UsersHandler {
	return &UsersHandler{apiConfig: apiConfig}
}

func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email string `json:"email"`
	}

	var userReq createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		log.Println("Failed to decode create user request", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	if userReq.Email == "" {
		log.Println("Empty email is not allowed")
		utils.BadRequestHandler(w, r)
		return
	}

	user, err := h.apiConfig.DBQueries.CreateUser(r.Context(), userReq.Email)
	if err != nil {
		log.Println("Failed to create new user", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	id, err := uuid.Parse(user.ID)
	if err != nil {
		log.Println("Invalid user ID returned from DB", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	result := User{
		ID:        id,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	log.Println("User created:", result)
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Println("Failed to marshal user", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	utils.RespondBytes(w, http.StatusCreated, bytes)
}
