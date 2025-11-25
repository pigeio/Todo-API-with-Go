package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid" // Ensure you have this: go get github.com/google/uuid
)

// TokenGenerator defines the behavior for auth token operations
type TokenGenerator interface {
	GenerateAccessToken(userID int, email string) (string, error)
	GenerateRefreshToken(userID int, email string) (string, string, error)
	ValidateToken(token string) (jwt.MapClaims, error)
	GetRefreshTokenExpiry() time.Duration
}

type JWTGenerator struct {
	SecretKey     []byte
	RefreshExpiry time.Duration
}

// NewJWTGenerator creates a new instance
func NewJWTGenerator(secret string) (*JWTGenerator, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET not set")
	}
	return &JWTGenerator{
		SecretKey:     []byte(secret),
		RefreshExpiry: 7 * 24 * time.Hour, // Default 7 days
	}, nil
}

// GenerateAccessToken creates a short-lived JWT
func (j *JWTGenerator) GenerateAccessToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(), // Short expiry for access token
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

// GenerateRefreshToken creates a long-lived JWT and returns the token + JTI
func (j *JWTGenerator) GenerateRefreshToken(userID int, email string) (string, string, error) {
	jti := uuid.New().String() // Unique identifier for this token
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"jti":   jti,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(j.RefreshExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.SecretKey)
	return signedToken, jti, err
}

// ValidateToken parses and validates a token string
func (j *JWTGenerator) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.SecretKey, nil
	})

	if err != nil || !t.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func (j *JWTGenerator) GetRefreshTokenExpiry() time.Duration {
	return j.RefreshExpiry
}
