package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pigeio/todo-api/internal/models"
)

type TodoRepository struct {
	db *pgxpool.Pool
}

func NewTodoRepository(db *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(ctx context.Context, todo *models.Todo) error {
	query := `
		INSERT INTO todos (user_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at, completed
	`

	err := r.db.QueryRow(ctx, query, todo.UserID, todo.Title, todo.Description).
		Scan(&todo.ID, &todo.CreatedAt, &todo.UpdatedAt, &todo.Completed)

	return err
}

func (r *TodoRepository) GetByID(ctx context.Context, id int) (*models.Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = $1
	`

	todo := &models.Todo{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("todo not found")
		}
		return nil, err
	}

	return todo, nil
}

func (r *TodoRepository) GetByUserID(ctx context.Context, userID, page, limit int, status, sortBy string) ([]models.Todo, int, error) {
	// 1. Build the base query and arguments
	var queryBuilder strings.Builder
	args := make([]interface{}, 0, 5) // Create a slice to hold our query arguments

	// Start with the base query for selecting todos
	queryBuilder.WriteString("SELECT id, user_id, title, description, completed, created_at, updated_at FROM todos WHERE user_id = $1")
	args = append(args, userID)
	argCounter := 2 // $1 is used for userID

	// 2. Add filters (status)
	if status == "completed" {
		queryBuilder.WriteString(fmt.Sprintf(" AND completed = $%d", argCounter))
		args = append(args, true)
		argCounter++
	} else if status == "pending" {
		queryBuilder.WriteString(fmt.Sprintf(" AND completed = $%d", argCounter))
		args = append(args, false)
		argCounter++
	}

	// 3. Get the Total Count *with* the filters applied
	// This is crucial for pagination. We run a COUNT on the filtered query.
	countQuery := "SELECT COUNT(*) FROM (" + queryBuilder.String() + ") AS filtered_todos"
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		log.Printf("Error counting todos: %v", err)
		return nil, 0, err
	}

	// 4. Add Sorting
	// We MUST whitelist sort_by values to prevent SQL injection.
	orderBy := "ORDER BY created_at DESC" // Default sort
	switch sortBy {
	case "title":
		orderBy = "ORDER BY title ASC"
	case "updated_at":
		orderBy = "ORDER BY updated_at DESC"
	}
	queryBuilder.WriteString(" " + orderBy) // It's safe to add this because it's from our whitelist

	// 5. Add Pagination
	offset := (page - 1) * limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1))
	args = append(args, limit, offset)

	// 6. Execute the final, dynamic query
	finalQuery := queryBuilder.String()
	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		log.Printf("Error querying todos: %v, query: %s", err, finalQuery)
		return nil, 0, err
	}
	defer rows.Close()

	// 7. Scan the results (no change from before)
	todos := []models.Todo{}
	for rows.Next() {
		var todo models.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		todos = append(todos, todo)
	}

	return todos, total, nil
}

func (r *TodoRepository) Update(ctx context.Context, todo *models.Todo) error {
	query := `
		UPDATE todos
		SET title = $1, description = $2, completed = $3, updated_at = NOW()
		WHERE id = $4 AND user_id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.ID,
		todo.UserID,
	).Scan(&todo.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("todo not found or unauthorized")
		}
		return err
	}

	return nil
}

func (r *TodoRepository) Delete(ctx context.Context, id, userID int) error {
	query := `DELETE FROM todos WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("todo not found or unauthorized")
	}

	return nil
}
