package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pigeio/todo-api/internal/models"
)

type RefreshRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshRepository(pool *pgxpool.Pool) *RefreshRepository {
	return &RefreshRepository{pool: pool}
}

func (r *RefreshRepository) CreateSession(ctx context.Context, s *models.RefreshSession) error {
	q := `INSERT INTO refresh_sessions (jti, user_id, user_agent, ip_address, created_at, expires_at)
	      VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(ctx, q, s.JTI, s.UserID, s.UserAgent, s.IP, s.CreatedAt, s.ExpiresAt)
	return err
}

func (r *RefreshRepository) GetSession(ctx context.Context, jti string) (*models.RefreshSession, error) {
	q := `SELECT jti, user_id, user_agent, ip_address, created_at, expires_at FROM refresh_sessions WHERE jti = $1`
	row := r.pool.QueryRow(ctx, q, jti)
	var s models.RefreshSession
	if err := row.Scan(&s.JTI, &s.UserID, &s.UserAgent, &s.IP, &s.CreatedAt, &s.ExpiresAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *RefreshRepository) DeleteSession(ctx context.Context, jti string) error {
	q := `DELETE FROM refresh_sessions WHERE jti = $1`
	_, err := r.pool.Exec(ctx, q, jti)
	return err
}

func (r *RefreshRepository) DeleteSessionsByUser(ctx context.Context, userID int) error {
	q := `DELETE FROM refresh_sessions WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, q, userID)
	return err
}
