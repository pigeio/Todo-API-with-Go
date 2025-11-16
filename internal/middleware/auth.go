package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pigeio/todo-api/internal/models" // Import models for Claims
	"github.com/pigeio/todo-api/internal/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware is now a function that ACCEPTS the tokenGenerator
// and RETURNS the actual middleware.
func AuthMiddleware(tokenGen utils.TokenGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			token := parts[1]

			// --- CHANGED ---
			// Use the injected token generator
			claims, err := tokenGen.ValidateToken(token)
			if err != nil {
				utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext is changed to use models.Claims
func GetUserFromContext(ctx context.Context) (*models.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*models.Claims)
	return claims, ok
}
