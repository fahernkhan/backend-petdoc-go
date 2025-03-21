package register

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Repository interface untuk mengelola operasi database
type Repository interface {
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CreateUser(ctx context.Context, user *User) error
}

type repository struct {
	db *sql.DB
}

// NewRepository membuat instance repository
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// CheckEmailExists memeriksa apakah email sudah digunakan
func (r *repository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 LIMIT 1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}

// CheckUsernameExists memeriksa apakah username sudah digunakan
func (r *repository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 LIMIT 1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}

// CreateUser menambahkan user baru ke dalam database
func (r *repository) CreateUser(ctx context.Context, user *User) error {
	query := `
        INSERT INTO users (
            full_name, email, password, gender, username, date_of_birth, role, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9
        )
        RETURNING id
    `

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.FullName,
		user.Email,
		user.Password,
		user.Gender,
		user.Username,
		user.DateOfBirth,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user insertion failed, no ID returned")
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}
