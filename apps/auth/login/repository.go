package login

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type loginRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &loginRepository{db: db}
}

func (r *loginRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
	SELECT 
            id, full_name, email, password, gender, username, date_of_birth, role, created_at, updated_at 
        FROM users 
        WHERE email = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Password,
		&user.Gender,
		&user.Username,
		&user.DateOfBirth,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
