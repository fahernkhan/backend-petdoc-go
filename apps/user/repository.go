package user

import (
	"context"
	"database/sql"
	"time"
)

type UserRepository interface {
	GetAllUsers(ctx context.Context, offset, limit int) ([]UserResponse, error)
	CountAllUsers(ctx context.Context) (int, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAllUsers(ctx context.Context, offset, limit int) ([]UserResponse, error) {
	query := `
		SELECT 
			id, 
			full_name, 
			email,
			gender,
			username,
			date_of_birth, 
			role
		FROM users 
		ORDER BY id ASC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, &DatabaseError{Operation: "query users", Err: err}
	}
	defer rows.Close()

	var users []UserResponse
	for rows.Next() {
		var u UserResponse
		var dateOfBirth time.Time // Tambahkan variabel untuk menangkap tanggal dari database

		err := rows.Scan(
			&u.ID,
			&u.FullName,
			&u.Email,
			&u.Gender,
			&u.Username,
			&dateOfBirth, // Ambil sebagai time.Time
			&u.Role,
		)
		if err != nil {
			return nil, &DatabaseError{Operation: "scan user", Err: err}
		}

		// Ubah format date_of_birth sebelum dimasukkan ke response
		u.DateOfBirth = dateOfBirth.Format("2006-01-02")

		users = append(users, u)
	}
	return users, nil
}

func (r *userRepository) CountAllUsers(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM users`

	var count int
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, &DatabaseError{Operation: "count users", Err: err}
	}

	return count, nil
}
