package users

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	cfg "github.com/dmchel/bootdev-chirpy/config"
	"github.com/dmchel/bootdev-chirpy/internal/auth"
	"github.com/dmchel/bootdev-chirpy/internal/database"
	"github.com/dmchel/bootdev-chirpy/utils"
	"github.com/google/uuid"
)

const DefaultExpiresInSeconds = 3600

type UsersHandler struct {
	apiConfig *cfg.ApiConfig
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func NewUserHandler(apiConfig *cfg.ApiConfig) *UsersHandler {
	return &UsersHandler{apiConfig: apiConfig}
}

func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	if userReq.Password == "" {
		log.Println("Empty password is not allowed")
		utils.BadRequestHandler(w, r)
		return
	}

	hash, err := auth.HashPassword(userReq.Password)
	if err != nil {
		log.Println("Failed to hash user password", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}
	var dbHash sql.NullString
	dbHash.String = hash
	dbHash.Valid = true

	user, err := h.apiConfig.DBQueries.CreateUser(r.Context(), database.CreateUserParams{Email: userReq.Email, HashedPassword: dbHash})
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

func (h *UsersHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	type loginUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var loginReq loginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		log.Println("Failed to decode login user request", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	if loginReq.Email == "" || loginReq.Password == "" {
		log.Println("Invlaid login or password")
		utils.UnauthorizedHandler(w, r)
		return
	}

	expiresInSeconds := DefaultExpiresInSeconds

	user, err := h.apiConfig.DBQueries.GetUser(r.Context(), loginReq.Email)
	if err != nil {
		log.Println("Failed to get user:", err)
		if err == sql.ErrNoRows {
			utils.UnauthorizedHandler(w, r)
		} else {
			utils.InternalServerErrorHandler(w, r)
		}
		return
	}

	if !user.HashedPassword.Valid {
		log.Println("User password is not set")
		utils.UnauthorizedHandler(w, r)
		return
	}

	match, err := auth.CheckPasswordHash(loginReq.Password, user.HashedPassword.String)
	if err != nil {
		log.Println("Couldn't check password hash", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}
	if !match {
		log.Println("Wrong password")
		utils.UnauthorizedHandler(w, r)
		return
	}

	id, err := uuid.Parse(user.ID)
	if err != nil {
		log.Println("Invalid user ID returned from DB", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	token, err := auth.MakeJWT(id, h.apiConfig.JWTSecret, time.Duration(expiresInSeconds)*time.Second)
	if err != nil {
		log.Println("Couldn't make JWT:", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	refreshToken, err := h.apiConfig.DBQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		log.Println("Couldn't save refresh token:", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	result := User{
		ID:           id,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken.Token,
	}

	log.Println("User login:", result)
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Println("Failed to marshal user", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	utils.RespondBytes(w, http.StatusOK, bytes)
}

func (h *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	type updateUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var userReq updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		log.Println("Failed to decode update user request", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	if userReq.Email == "" {
		log.Println("Empty email is not allowed")
		utils.BadRequestHandler(w, r)
		return
	}

	if userReq.Password == "" {
		log.Println("Empty password is not allowed")
		utils.BadRequestHandler(w, r)
		return
	}

	authUserId, err := utils.GetUserIdentityFromAuth(r, h.apiConfig.JWTSecret)
	if err != nil {
		log.Println("Auth failed", err)
		utils.UnauthorizedHandler(w, r)
		return
	}

	hash, err := auth.HashPassword(userReq.Password)
	if err != nil {
		log.Println("Failed to hash user password", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}
	var dbHash sql.NullString
	dbHash.String = hash
	dbHash.Valid = true

	user, err := h.apiConfig.DBQueries.UpdateUser(r.Context(), database.UpdateUserParams{Email: userReq.Email, HashedPassword: dbHash, ID: authUserId.String()})
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

	log.Println("User updated:", result)
	bytes, err := json.Marshal(result)
	if err != nil {
		log.Println("Failed to marshal user", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	utils.RespondBytes(w, http.StatusOK, bytes)
}
