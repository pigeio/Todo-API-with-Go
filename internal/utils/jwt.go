package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pigeio/todo-api/internal/models" // We moved Claims here
)

// TokenGenerator is the "plug socket" (interface) for our token generator.
type TokenGenerator interface {
	GenerateToken(userID int, email string) (string, error)
	ValidateToken(tokenString string) (*models.Claims, error)
}

// JWTGenerator is our REAL implementation that fits the socket
type JWTGenerator struct {
	SecretKey []byte
}

// NewJWTGenerator creates the REAL generator (we'll call this in main.go)
func NewJWTGenerator(secret string) (*JWTGenerator, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET not set")
	}
	return &JWTGenerator{SecretKey: []byte(secret)}, nil
}

// GenerateToken implements the TokenGenerator interface
func (j *JWTGenerator) GenerateToken(userID int, email string) (string, error) {
	claims := models.Claims{ // Now reads from models
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

// ValidateToken implements the TokenGenerator interface
func (j *JWTGenerator) ValidateToken(tokenString string) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
