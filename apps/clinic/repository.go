package clinic

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, clinic *ClinicResponse) error
	GetByID(ctx context.Context, id int) (ClinicResponse, error)
	GetAll(ctx context.Context, page, pageSize int) ([]ClinicResponse, int64, error)
	Update(ctx context.Context, clinic *ClinicResponse) error
	Delete(ctx context.Context, id, userID int) error
	Search(ctx context.Context, query string, page, pageSize int) ([]ClinicResponse, int64, error)
}

type clinicRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &clinicRepo{db: db}
}

func (r *clinicRepo) Create(ctx context.Context, clinic *ClinicResponse) error {
	query := `INSERT INTO clinics (
		name, phone_number, address, map_link, image, user_id
	) VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		clinic.Name,
		clinic.PhoneNumber,
		clinic.Address,
		clinic.MapLink,
		clinic.ImageURL,
		clinic.UserID,
	).Scan(&clinic.ID, &clinic.CreatedAt, &clinic.UpdatedAt)
}

func (r *clinicRepo) GetByID(ctx context.Context, id int) (ClinicResponse, error) {
	query := `SELECT id, name, phone_number, address, map_link, image, user_id, 
		created_at, updated_at FROM clinics WHERE id = $1`

	var clinic ClinicResponse
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&clinic.ID,
		&clinic.Name,
		&clinic.PhoneNumber,
		&clinic.Address,
		&clinic.MapLink,
		&clinic.ImageURL,
		&clinic.UserID,
		&clinic.CreatedAt,
		&clinic.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return ClinicResponse{}, ErrClinicNotFound
		}
		return ClinicResponse{}, err
	}

	return clinic, nil
}

func (r *clinicRepo) GetAll(ctx context.Context, page, pageSize int) ([]ClinicResponse, int64, error) {
	offset := (page - 1) * pageSize

	query := `SELECT id, name, phone_number, address, map_link, image, user_id,
		created_at, updated_at FROM clinics
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clinics []ClinicResponse
	for rows.Next() {
		var clinic ClinicResponse
		if err := rows.Scan(
			&clinic.ID,
			&clinic.Name,
			&clinic.PhoneNumber,
			&clinic.Address,
			&clinic.MapLink,
			&clinic.ImageURL,
			&clinic.UserID,
			&clinic.CreatedAt,
			&clinic.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		clinics = append(clinics, clinic)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM clinics`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	return clinics, total, nil
}

func (r *clinicRepo) Update(ctx context.Context, clinic *ClinicResponse) error {
	query := `UPDATE clinics SET
		name = $1,
		phone_number = $2,
		address = $3,
		map_link = $4,
		image = $5,
		updated_at = NOW()
		WHERE id = $6 AND user_id = $7
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		clinic.Name,
		clinic.PhoneNumber,
		clinic.Address,
		clinic.MapLink,
		clinic.ImageURL,
		clinic.ID,
		clinic.UserID,
	).Scan(&clinic.UpdatedAt)
}

func (r *clinicRepo) Delete(ctx context.Context, id, userID int) error {
	query := `DELETE FROM clinics WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrClinicNotFound
	}

	return nil
}

func (r *clinicRepo) Search(ctx context.Context, query string, page, pageSize int) ([]ClinicResponse, int64, error) {
	offset := (page - 1) * pageSize
	searchQuery := fmt.Sprintf("%%%s%%", query)

	sqlQuery := `SELECT id, name, phone_number, address, map_link, image, user_id,
		created_at, updated_at FROM clinics
		WHERE name ILIKE $1 OR address ILIKE $1 OR phone_number ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, sqlQuery, searchQuery, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clinics []ClinicResponse
	for rows.Next() {
		var clinic ClinicResponse
		if err := rows.Scan(
			&clinic.ID,
			&clinic.Name,
			&clinic.PhoneNumber,
			&clinic.Address,
			&clinic.MapLink,
			&clinic.ImageURL,
			&clinic.UserID,
			&clinic.CreatedAt,
			&clinic.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		clinics = append(clinics, clinic)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM clinics 
		WHERE name ILIKE $1 OR address ILIKE $1 OR phone_number ILIKE $1`
	if err := r.db.QueryRowContext(ctx, countQuery, searchQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	return clinics, total, nil
}
