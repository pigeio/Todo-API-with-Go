package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pigeio/todo-api/internal/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(tokenGen utils.TokenGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := tokenGen.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			subVal, ok := claims["sub"]
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID := int(subVal.(float64))

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}
