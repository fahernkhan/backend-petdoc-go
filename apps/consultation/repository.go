package consultation

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository interface {
	CreateConsultation(ctx context.Context, cons *ConsultationResponse) error
	GetConsultations(ctx context.Context, userID, page, pageSize int) ([]ConsultationResponse, int64, error)
	CheckDoctorAvailability(ctx context.Context, doctorID int, start, end time.Time) (bool, error)
	GetDoctorDetails(ctx context.Context, doctorID int) (gmeetLink string, pricePerHour float64, err error)
}

type consultationRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &consultationRepo{db: db}
}

func (r *consultationRepo) CreateConsultation(ctx context.Context, cons *ConsultationResponse) error {
	query := `
	INSERT INTO consultations (
		user_id, doctor_id, pet_type, pet_name, pet_age,
		disease_description, consultation_date, start_time,
		end_time, payment_proof
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		cons.UserID,
		cons.DoctorID,
		cons.PetType,
		cons.PetName,
		cons.PetAge,
		cons.DiseaseDescription,
		cons.ConsultationDate,
		cons.StartTime,
		cons.EndTime,
		cons.PaymentProof,
	).Scan(&cons.ID, &cons.CreatedAt)
}

func (r *consultationRepo) GetConsultations(ctx context.Context, userID, page, pageSize int) ([]ConsultationResponse, int64, error) {
	offset := (page - 1) * pageSize
	query := `
	SELECT id, user_id, doctor_id, pet_type, pet_name, pet_age,
		   disease_description, consultation_date, start_time,
		   end_time, payment_proof, created_at
	FROM consultations
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var consultations []ConsultationResponse
	for rows.Next() {
		var c ConsultationResponse
		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.DoctorID,
			&c.PetType,
			&c.PetName,
			&c.PetAge,
			&c.DiseaseDescription,
			&c.ConsultationDate,
			&c.StartTime,
			&c.EndTime,
			&c.PaymentProof,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		consultations = append(consultations, c)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM consultations WHERE user_id = $1`
	err = r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)

	return consultations, total, err
}

func (r *consultationRepo) CheckDoctorAvailability(ctx context.Context, doctorID int, start, end time.Time) (bool, error) {
	fmt.Printf("Checking availability for doctor %d between %s and %s\n",
		doctorID,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339))

	// ... sisa kode
	query := `
    SELECT NOT EXISTS(
        SELECT 1 FROM consultations 
        WHERE doctor_id = $1 
        AND (start_time, end_time) OVERLAPS ($2, $3)
    )` // <-- Tambahkan tanda kurung penutup

	var available bool
	err := r.db.QueryRowContext(ctx, query, doctorID, start, end).Scan(&available)

	if err != nil {
		return false, fmt.Errorf("error checking availability: %w", err)
	}

	return available, nil
}

func (r *consultationRepo) GetDoctorDetails(ctx context.Context, doctorID int) (string, float64, error) {
	var gmeetLink string
	var price float64
	query := `SELECT gmeet_link, price_per_hour FROM doctors WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, doctorID).Scan(&gmeetLink, &price)
	return gmeetLink, price, err
}
