package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pigeio/todo-api/internal/models"
	"github.com/pigeio/todo-api/internal/repository" // Import for the interface
	"github.com/pigeio/todo-api/internal/utils"      // Import for the interface
)

type AuthHandler struct {
	// We now use the interfaces (the "sockets")
	userRepo  repository.User_Repository // Using your name from interfaces.go
	validator *validator.Validate
	tokenGen  utils.TokenGenerator // From the new jwt.go
}

// NewAuthHandler now accepts the interfaces
func NewAuthHandler(userRepo repository.User_Repository, tokenGen utils.TokenGenerator) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		validator: validator.New(),
		tokenGen:  tokenGen, // Store the token generator
	}
}

// --- Register Handler (Updated) ---
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		// You can add your detailed error formatting back here
		utils.RespondError(w, http.StatusBadRequest, "Validation failed")
		return
	}

	exists, err := h.userRepo.EmailExists(r.Context(), req.Email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if exists {
		utils.RespondError(w, http.StatusBadRequest, "Email already exists")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	user := &models.User{
		Name:     req.Name,
		Email:    strings.ToLower(req.Email),
		Password: hashedPassword,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate token (now uses the interface)
	token, err := h.tokenGen.GenerateToken(user.ID, user.Email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := models.AuthResponse{Token: token}
	utils.RespondJSON(w, http.StatusCreated, response)
}

// --- Login Handler (Updated) ---
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Validation failed")
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), strings.ToLower(req.Email))
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate token (now uses the interface)
	token, err := h.tokenGen.GenerateToken(user.ID, user.Email)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := models.AuthResponse{Token: token}
	utils.RespondJSON(w, http.StatusOK, response)
}
