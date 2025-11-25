package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pigeio/todo-api/internal/models"
	"github.com/pigeio/todo-api/internal/repository"
	"github.com/pigeio/todo-api/internal/utils"
)

type RefreshHandler struct {
	UserRepo    repository.User_Repository
	RefreshRepo *repository.RefreshRepository
	TokenGen    utils.TokenGenerator // FIX: Inject the generator
}

// Update constructor to accept TokenGenerator
func NewRefreshHandler(u repository.User_Repository, r *repository.RefreshRepository, tg utils.TokenGenerator) *RefreshHandler {
	return &RefreshHandler{
		UserRepo:    u,
		RefreshRepo: r,
		TokenGen:    tg,
	}
}

var (
	refreshCookieName = func() string {
		n := os.Getenv("REFRESH_COOKIE_NAME")
		if n == "" {
			return "refreshToken"
		}
		return n
	}()
	cookieSecure = os.Getenv("COOKIE_SECURE") == "true"
)

func (h *RefreshHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var tokenStr string

	if c, err := r.Cookie(refreshCookieName); err == nil {
		tokenStr = c.Value
	} else {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		tokenStr = body.RefreshToken
	}

	if tokenStr == "" {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	// FIX: Use injected TokenGen
	claims, err := h.TokenGen.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	jti, _ := claims["jti"].(string)
	subVal := claims["sub"]

	// Handle "sub" being float64 (JSON default) or string
	var userID int
	switch v := subVal.(type) {
	case float64:
		userID = int(v)
	case string:
		userID, _ = strconv.Atoi(v)
	case int:
		userID = v
	}

	if jti == "" || userID == 0 {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	session, err := h.RefreshRepo.GetSession(r.Context(), jti)
	if err != nil || session == nil {
		http.Error(w, "refresh session not found or revoked", http.StatusUnauthorized)
		return
	}

	if session.ExpiresAt.Before(time.Now().UTC()) {
		_ = h.RefreshRepo.DeleteSession(r.Context(), jti)
		http.Error(w, "refresh token expired", http.StatusUnauthorized)
		return
	}

	_ = h.RefreshRepo.DeleteSession(r.Context(), jti)

	// FIX: Use injected TokenGen
	accessToken, err := h.TokenGen.GenerateAccessToken(userID, "")
	if err != nil {
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	newRefreshToken, newJti, err := h.TokenGen.GenerateRefreshToken(userID, "")
	if err != nil {
		http.Error(w, "failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	newSession := &models.RefreshSession{
		JTI:       newJti,
		UserID:    userID,
		UserAgent: r.UserAgent(),
		IP:        r.RemoteAddr,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().Add(h.TokenGen.GetRefreshTokenExpiry()),
	}

	if err := h.RefreshRepo.CreateSession(r.Context(), newSession); err != nil {
		http.Error(w, "failed to save refresh session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   cookieSecure,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  newSession.ExpiresAt,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})
}

func (h *RefreshHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(refreshCookieName); err == nil {
		// FIX: Use injected TokenGen
		claims, err := h.TokenGen.ValidateToken(c.Value)
		if err == nil {
			if jti, ok := claims["jti"].(string); ok {
				_ = h.RefreshRepo.DeleteSession(r.Context(), jti)
			}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   cookieSecure,
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	w.Write([]byte(`{"message": "logged out"}`))
}

func (h *RefreshHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value("user_id")
	if value == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	switch v := value.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	case string:
		userID, _ = strconv.Atoi(v)
	default:
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	_ = h.RefreshRepo.DeleteSessionsByUser(r.Context(), userID)

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   cookieSecure,
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	w.Write([]byte(`{"message": "logged out from all devices"}`))
}
