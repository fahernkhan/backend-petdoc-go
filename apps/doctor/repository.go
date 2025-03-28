package doctor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	UpdateUserRole(ctx context.Context, tx *sql.Tx, userID int, newRole string) error
	GetUserRole(ctx context.Context, tx *sql.Tx, userID int) (string, error)
	Create(ctx context.Context, tx *sql.Tx, d *DoctorRequest) (int, error)
	GetByIDWithTx(ctx context.Context, tx *sql.Tx, id int) (DoctorResponse, error)
	GetByID(ctx context.Context, id int) (DoctorResponse, error)
	GetAll(ctx context.Context, page, limit int) ([]DoctorResponse, int, error)
	Update(ctx context.Context, id int, doctor *DoctorRequest) error
	Delete(ctx context.Context, id int) error
	DeleteWithTx(ctx context.Context, tx *sql.Tx, id int) error
}

type repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}

func (r *repo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// Di repository.go - perbaiki error handling
func (r *repo) UpdateUserRole(ctx context.Context, tx *sql.Tx, userID int, newRole string) error {
	result, err := tx.ExecContext(ctx,
		`UPDATE users SET role = $1 WHERE id = $2`,
		newRole,
		userID,
	)

	if err != nil {
		// Handle constraint violation
		if strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("%w: %v", ErrDuplicateEntry, err)
		}
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *repo) GetUserRole(ctx context.Context, tx *sql.Tx, userID int) (string, error) {
	query := `SELECT role FROM users WHERE id = $1`
	var role string
	err := tx.QueryRowContext(ctx, query, userID).Scan(&role)
	return role, err
}

// Tambahkan parameter tx ke semua method
func (r *repo) Create(ctx context.Context, tx *sql.Tx, d *DoctorRequest) (int, error) {
	workingDays, _ := json.Marshal(d.WorkingDays)
	workingHours, _ := json.Marshal(map[string]string{
		"start": d.WorkingHoursStart,
		"end":   d.WorkingHoursEnd,
	})

	query := `INSERT INTO doctors (
        user_id, full_name, last_education, specialist_at, 
        profile_image, birth_date, hospital_name, years_of_experience,
        price_per_hour, gmeet_link, working_days, working_hours
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    RETURNING id`

	var id int
	err := tx.QueryRowContext(ctx, query,
		d.UserID,
		d.FullName,
		d.LastEducation,
		d.SpecialistAt,
		d.ProfileImage,
		d.BirthDate,
		d.HospitalName,
		d.YearsOfExperience,
		d.PricePerHour,
		d.GmeetLink,
		workingDays,
		workingHours,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}
	return id, nil
}

// Implementasi method lainnya dengan pola yang sama
// GetByID implementation
func (r *repo) GetByIDWithTx(ctx context.Context, tx *sql.Tx, id int) (DoctorResponse, error) {
	query := `
    SELECT id, user_id, full_name, last_education, specialist_at, profile_image, 
           birth_date, hospital_name, years_of_experience, price_per_hour, 
           gmeet_link, working_days, working_hours, created_at, updated_at 
    FROM doctors 
    WHERE id = $1`

	var d DoctorResponse
	var workingDays []byte
	var workingHours []byte
	var birthDate time.Time

	err := tx.QueryRowContext(ctx, query, id).Scan(
		&d.ID,
		&d.UserID,
		&d.FullName,
		&d.LastEducation,
		&d.SpecialistAt,
		&d.ProfileImage,
		&birthDate,
		&d.HospitalName,
		&d.YearsOfExperience,
		&d.PricePerHour,
		&d.GmeetLink,
		&workingDays,
		&workingHours,
		&d.CreatedAt,
		&d.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return DoctorResponse{}, ErrDoctorNotFound
		}
		return DoctorResponse{}, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	// Parse JSON fields
	json.Unmarshal(workingDays, &d.WorkingDays)
	json.Unmarshal(workingHours, &d.WorkingHours)
	d.BirthDate = birthDate.Format("2006-01-02")

	return d, nil
}

// Implementasi method lainnya dengan pola yang sama
// GetByID implementation
func (r *repo) GetByID(ctx context.Context, id int) (DoctorResponse, error) {
	query := `
    SELECT id, user_id, full_name, last_education, specialist_at, profile_image, 
           birth_date, hospital_name, years_of_experience, price_per_hour, 
           gmeet_link, working_days, working_hours, created_at, updated_at 
    FROM doctors 
    WHERE id = $1`

	var d DoctorResponse
	var workingDays []byte
	var workingHours []byte
	var birthDate time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID,
		&d.UserID,
		&d.FullName,
		&d.LastEducation,
		&d.SpecialistAt,
		&d.ProfileImage,
		&birthDate,
		&d.HospitalName,
		&d.YearsOfExperience,
		&d.PricePerHour,
		&d.GmeetLink,
		&workingDays,
		&workingHours,
		&d.CreatedAt,
		&d.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return DoctorResponse{}, ErrDoctorNotFound
		}
		return DoctorResponse{}, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	// Parse JSON fields
	json.Unmarshal(workingDays, &d.WorkingDays)
	json.Unmarshal(workingHours, &d.WorkingHours)
	d.BirthDate = birthDate.Format("2006-01-02")

	return d, nil
}

// GetAll implementation with pagination
func (r *repo) GetAll(ctx context.Context, page, limit int) ([]DoctorResponse, int, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM doctors`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	// Get paginated data
	query := `
    SELECT id, user_id, full_name, last_education, specialist_at, profile_image, 
           birth_date, hospital_name, years_of_experience, price_per_hour, 
           gmeet_link, working_days, working_hours, created_at, updated_at 
    FROM doctors 
    ORDER BY id DESC 
    LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}
	defer rows.Close()

	var doctors []DoctorResponse
	for rows.Next() {
		var d DoctorResponse
		var workingDays []byte
		var workingHours []byte
		var birthDate time.Time

		err := rows.Scan(
			&d.ID,
			&d.UserID,
			&d.FullName,
			&d.LastEducation,
			&d.SpecialistAt,
			&d.ProfileImage,
			&birthDate,
			&d.HospitalName,
			&d.YearsOfExperience,
			&d.PricePerHour,
			&d.GmeetLink,
			&workingDays,
			&workingHours,
			&d.CreatedAt,
			&d.UpdatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		}

		json.Unmarshal(workingDays, &d.WorkingDays)
		json.Unmarshal(workingHours, &d.WorkingHours)
		d.BirthDate = birthDate.Format("2006-01-02")
		doctors = append(doctors, d)
	}

	return doctors, total, nil
}

// Update implementation
func (r *repo) Update(ctx context.Context, id int, d *DoctorRequest) error {
	workingDays, _ := json.Marshal(d.WorkingDays)
	workingHours, _ := json.Marshal(map[string]string{
		"start": d.WorkingHoursStart,
		"end":   d.WorkingHoursEnd,
	})

	query := `
    UPDATE doctors 
    SET full_name = $1,
        last_education = $2,
        specialist_at = $3,
        profile_image = $4,
        birth_date = $5,
        hospital_name = $6,
        years_of_experience = $7,
        price_per_hour = $8,
        gmeet_link = $9,
        working_days = $10,
        working_hours = $11,
        updated_at = NOW()
    WHERE id = $12`

	result, err := r.db.ExecContext(ctx, query,
		d.FullName,
		d.LastEducation,
		d.SpecialistAt,
		d.ProfileImage,
		d.BirthDate,
		d.HospitalName,
		d.YearsOfExperience,
		d.PricePerHour,
		d.GmeetLink,
		workingDays,
		workingHours,
		id,
	)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrDoctorNotFound
	}

	return nil
}

// Delete implementation
func (r *repo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM doctors WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrDoctorNotFound
	}

	return nil
}

func (r *repo) DeleteWithTx(ctx context.Context, tx *sql.Tx, id int) error {
	query := `DELETE FROM doctors WHERE id = $1`
	result, err := tx.ExecContext(ctx, query, id)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrDoctorNotFound
	}

	return nil
}
