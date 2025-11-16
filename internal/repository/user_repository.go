package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pigeio/todo-api/internal/models"
)

// UserRepository is the REAL struct that holds the database connection
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository
// This returns the STRUCT, which perfectly matches the User_Repository INTERFACE
func NewUserRepository(db *pgxpool.Pool) User_Repository {
	return &UserRepository{db: db}
}

// Create implements the User_Repository interface
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, user.Name, user.Email, user.Password).
		Scan(&user.ID, &user.CreatedAt)

	return err
}

// GetByEmail implements the User_Repository interface
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password, created_at
		FROM users
		WHERE email = $1
	`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// EmailExists implements the User_Repository interface
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
