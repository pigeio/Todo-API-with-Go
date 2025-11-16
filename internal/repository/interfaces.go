package repository

import (
	"context"

	"github.com/pigeio/todo-api/internal/models"
)

// UserRepository defines the interface for user-related database operations
type User_Repository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

// TodoRepository defines the interface for todo-related database operations
type Todo_Repository interface {
	Create(ctx context.Context, todo *models.Todo) error
	GetByID(ctx context.Context, id int) (*models.Todo, error)
	GetByUserID(ctx context.Context, userID, page, limit int, status, sortBy string) ([]models.Todo, int, error)
	Update(ctx context.Context, todo *models.Todo) error
	Delete(ctx context.Context, id, userID int) error
}
