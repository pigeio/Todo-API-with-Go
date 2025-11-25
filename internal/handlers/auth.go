package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pigeio/todo-api/internal/models"
	"github.com/pigeio/todo-api/internal/repository"
	"github.com/pigeio/todo-api/internal/utils"
)

type AuthHandler struct {
	userRepo    repository.User_Repository
	validator   *validator.Validate
	tokenGen    utils.TokenGenerator
	refreshRepo *repository.RefreshRepository
}

// NewAuthHandler (for tests)
func NewAuthHandler(
	userRepo repository.User_Repository,
	tokenGen utils.TokenGenerator,
) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		validator: validator.New(),
		tokenGen:  tokenGen,
	}
}

// NewAuthHandlerWithRefresh (for main app)
func NewAuthHandlerWithRefresh(
	userRepo repository.User_Repository,
	tokenGen utils.TokenGenerator,
	refreshRepo *repository.RefreshRepository,
) *AuthHandler {
	h := NewAuthHandler(userRepo, tokenGen)
	h.refreshRepo = refreshRepo
	return h
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.validator.Struct(req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Validation failed")
		return
	}

	exists, _ := h.userRepo.EmailExists(r.Context(), req.Email)
	if exists {
		utils.RespondError(w, http.StatusBadRequest, "Email already exists")
		return
	}

	hashed, _ := utils.HashPassword(req.Password)

	user := &models.User{
		Name:     req.Name,
		Email:    strings.ToLower(req.Email),
		Password: hashed,
	}

	h.userRepo.Create(r.Context(), user)

	// Use method from interface
	token, _ := h.tokenGen.GenerateAccessToken(user.ID, user.Email)

	utils.RespondJSON(w, http.StatusCreated, map[string]string{
		"token": token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	user, err := h.userRepo.GetByEmail(r.Context(), strings.ToLower(req.Email))
	if err != nil || !utils.CheckPassword(req.Password, user.Password) {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// MODE 1 — no refresh repo → tests (simple token)
	if h.refreshRepo == nil {
		token, _ := h.tokenGen.GenerateAccessToken(user.ID, user.Email)
		utils.RespondJSON(w, http.StatusOK, map[string]string{"token": token})
		return
	}

	// MODE 2 — refresh tokens (real app)
	// FIX: Call interface methods instead of static utils.* functions
	accessToken, _ := h.tokenGen.GenerateAccessToken(user.ID, user.Email)
	refreshToken, jti, _ := h.tokenGen.GenerateRefreshToken(user.ID, user.Email)

	session := &models.RefreshSession{
		JTI:       jti,
		UserID:    user.ID,
		UserAgent: r.UserAgent(),
		IP:        r.RemoteAddr,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().Add(h.tokenGen.GetRefreshTokenExpiry()), // FIX
	}
	h.refreshRepo.CreateSession(r.Context(), session)

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}
