package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pigeio/todo-api/internal/models"
)

// --- 1. The Mock Repository ---
// This mock satisfies the repository.User_Repository interface
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

// --- 2. The NEW Mock Token Generator ---
// This mock satisfies the utils.TokenGenerator interface
type MockTokenGenerator struct{}

func (m *MockTokenGenerator) GenerateToken(userID int, email string) (string, error) {
	return "mock_token_string", nil
}
func (m *MockTokenGenerator) ValidateToken(tokenString string) (*models.Claims, error) {
	return nil, nil // Not needed for this test
}

// --- 3. The Test Function (FIXED) ---
func TestRegisterHandler(t *testing.T) {

	t.Run("successful registration", func(t *testing.T) {
		// A. Setup
		registerReq := models.RegisterRequest{
			Name: "Test User", Email: "test@example.com", Password: "password123",
		}
		body, _ := json.Marshal(registerReq)

		// Create our mocks
		mockRepo := &MockUserRepository{
			MockEmailExists: func(ctx context.Context, email string) (bool, error) {
				return false, nil // "Email does not exist"
			},
			MockCreate: func(ctx context.Context, user *models.User) error {
				user.ID = 1 // Simulate assigning an ID
				return nil
			},
		}
		mockTokenGen := &MockTokenGenerator{} // Our new mock

		// --- THIS IS THE FIX ---
		handler := NewAuthHandler(mockRepo, mockTokenGen)

		// B. Create request/recorder
		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		// C. Run
		handler.Register(rr, req)

		// D. Check
		if rr.Code != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v",
				rr.Code, http.StatusCreated)
		}
		if !strings.Contains(rr.Body.String(), "mock_token_string") {
			t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
		}
	})

	t.Run("email already exists", func(t *testing.T) {
		registerReq := models.RegisterRequest{
			Name: "Test User", Email: "test@example.com", Password: "password123",
		}
		body, _ := json.Marshal(registerReq)

		mockRepo := &MockUserRepository{
			MockEmailExists: func(ctx context.Context, email string) (bool, error) {
				return true, nil // "Email *does* exist"
			},
		}
		mockTokenGen := &MockTokenGenerator{} // Not used, but required

		handler := NewAuthHandler(mockRepo, mockTokenGen)
		req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handler.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				rr.Code, http.StatusBadRequest)
		}
		if !strings.Contains(rr.Body.String(), "Email already exists") {
			t.Errorf("handler returned wrong error message: got %v", rr.Body.String())
		}
	})
}
