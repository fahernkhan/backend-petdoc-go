package admin

import (
	"context"
	"database/sql"
	"fmt"
)

type AdminRepository interface {
	GetUserRole(ctx context.Context, userID int) (string, error)
	UpdateUserRole(ctx context.Context, userID int, newRole string) error
}

type adminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetUserRole(ctx context.Context, userID int) (string, error) {

	query := `
		SELECT role 
		FROM users 
		WHERE id = $1`

	fmt.Printf("Executing query: %s\nParams: %v\n", query, userID)

	var role string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&role)

	if err == sql.ErrNoRows {
		fmt.Printf("Database error: %v\n", err) // Ini akan muncul di console server
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return role, nil
}

func (r *adminRepository) UpdateUserRole(ctx context.Context, userID int, newRole string) error {
	query := `
        UPDATE users 
        SET role = $1, updated_at = NOW() 
        WHERE id = $2`

	fmt.Printf("Executing query: %s\nParams: %v\n", query, userID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, query, newRole, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	return nil
}
