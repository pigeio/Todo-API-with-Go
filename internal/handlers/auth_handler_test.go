package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5" // Import jwt
	"github.com/pigeio/todo-api/internal/models"
)

// 1. Mock User Repository (Unchanged)
type MockUserRepository struct {
	MockEmailExists func(ctx context.Context, email string) (bool, error)
	MockCreate      func(ctx context.Context, user *models.User) error
	MockGetByEmail  func(ctx context.Context, email string) (*models.User, error)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return m.MockEmailExists(ctx, email)
}
func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	return m.MockCreate(ctx, user)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.MockGetByEmail(ctx, email)
}

// 2. Mock Token Generator (UPDATED to match interface)
type MockTokenGenerator struct{}

func (m *MockTokenGenerator) GenerateAccessToken(userID int, email string) (string, error) {
	return "mock_access_token", nil
}

// Added missing method
func (m *MockTokenGenerator) GenerateRefreshToken(userID int, email string) (string, string, error) {
	return "mock_refresh_token", "mock_jti", nil
}

// Fixed return type: returns jwt.MapClaims instead of *models.Claims
func (m *MockTokenGenerator) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	return jwt.MapClaims{
		"sub": 1,
		"jti": "mock_jti",
	}, nil
}

// Added missing method
func (m *MockTokenGenerator) GetRefreshTokenExpiry() time.Duration {
	return 24 * time.Hour
}

// 3. Mock Refresh Repository (Unchanged)
type MockRefreshRepository struct{}

func (m *MockRefreshRepository) CreateSession(ctx context.Context, s *models.RefreshSession) error {
	return nil
}
func (m *MockRefreshRepository) GetSession(ctx context.Context, jti string) (*models.RefreshSession, error) {
	return nil, nil
}
func (m *MockRefreshRepository) DeleteSession(ctx context.Context, jti string) error {
	return nil
}
func (m *MockRefreshRepository) DeleteSessionsByUser(ctx context.Context, uid int) error {
	return nil
}

// 4. TESTS
func TestRegisterHandler(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		registerReq := models.RegisterRequest{
			Name: "Test User", Email: "test@example.com", Password: "password123",
		}
		body, _ := json.Marshal(registerReq)

		mockRepo := &MockUserRepository{
			MockEmailExists: func(ctx context.Context, email string) (bool, error) {
				return false, nil
			},
			MockCreate: func(ctx context.Context, user *models.User) error {
				user.ID = 1
				return nil
			},
		}

		mockTokenGen := &MockTokenGenerator{}
		handler := NewAuthHandler(mockRepo, mockTokenGen)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("status got=%v want=%v", rr.Code, http.StatusCreated)
		}

		// Expect mock token string
		if !strings.Contains(rr.Body.String(), "mock_access_token") {
			t.Errorf("unexpected body: %v", rr.Body.String())
		}
	})

	t.Run("email already exists", func(t *testing.T) {
		registerReq := models.RegisterRequest{
			Name: "Test User", Email: "test@example.com", Password: "password123",
		}
		body, _ := json.Marshal(registerReq)

		mockRepo := &MockUserRepository{
			MockEmailExists: func(ctx context.Context, email string) (bool, error) {
				return true, nil
			},
		}

		mockTokenGen := &MockTokenGenerator{}
		handler := NewAuthHandler(mockRepo, mockTokenGen)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status got=%v want=%v", rr.Code, http.StatusBadRequest)
		}

		if !strings.Contains(rr.Body.String(), "Email already exists") {
			t.Errorf("wrong message: %v", rr.Body.String())
		}
	})
}
