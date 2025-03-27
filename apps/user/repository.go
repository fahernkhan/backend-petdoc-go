package user

import (
	"context"
	"database/sql"
	"time"
)

type UserRepository interface {
	GetAllUsers(ctx context.Context, offset, limit int, filter string) ([]UserResponse, error)
	CountAllUsers(ctx context.Context, filter string) (int, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAllUsers(ctx context.Context, offset, limit int, filter string) ([]UserResponse, error) {
	baseQuery := `
        SELECT 
            id, 
            full_name, 
            email,
            gender,
            username,
            date_of_birth, 
            role
        FROM users`

	args := []interface{}{limit, offset}
	whereClause := ""

	if filter != "" {
		whereClause = " WHERE (full_name ILIKE $3 OR email ILIKE $3 OR username ILIKE $3)"
		args = append(args, "%"+filter+"%")
	}

	query := baseQuery + whereClause + " ORDER BY id ASC LIMIT $1 OFFSET $2"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, &DatabaseError{Operation: "query users", Err: err}
	}
	defer rows.Close()

	var users []UserResponse
	for rows.Next() {
		var u UserResponse
		var dateOfBirth time.Time

		err := rows.Scan(
			&u.ID,
			&u.FullName,
			&u.Email,
			&u.Gender,
			&u.Username,
			&dateOfBirth,
			&u.Role,
		)
		if err != nil {
			return nil, &DatabaseError{Operation: "scan user", Err: err}
		}

		u.DateOfBirth = dateOfBirth.Format("2006-01-02")
		users = append(users, u)
	}
	return users, nil
}

func (r *userRepository) CountAllUsers(ctx context.Context, filter string) (int, error) {
	baseQuery := `SELECT COUNT(*) FROM users`
	args := []interface{}{}

	if filter != "" {
		baseQuery += " WHERE (full_name ILIKE $1 OR email ILIKE $1 OR username ILIKE $1)"
		args = append(args, "%"+filter+"%")
	}

	var count int
	if err := r.db.QueryRowContext(ctx, baseQuery, args...).Scan(&count); err != nil {
		return 0, &DatabaseError{Operation: "count users", Err: err}
	}
	return count, nil
}
