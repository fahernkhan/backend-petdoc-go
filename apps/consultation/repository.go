package consultation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Repository interface {
	CreateConsultation(ctx context.Context, cons *ConsultationResponse) error
	GetConsultations(ctx context.Context, userID, page, pageSize int) ([]ConsultationResponse, int64, error)
	CheckDoctorAvailability(ctx context.Context, doctorID int, start, end time.Time) (bool, error)
	GetDoctorDetails(ctx context.Context, doctorID int) (DoctorSchedule, error)
	CheckAvailability(ctx context.Context, doctorID, userID int, start, end time.Time) (bool, error)
	DoctorExists(ctx context.Context, doctorID int) (bool, error)
}

type consultationRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &consultationRepo{db: db}
}

func (r *consultationRepo) CreateConsultation(ctx context.Context, cons *ConsultationResponse) error {
	// Parse string ke time.Time
	startTime, err := time.Parse(time.RFC3339, cons.StartTimeUTC)
	if err != nil {
		return fmt.Errorf("invalid start_time format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, cons.EndTimeUTC)
	if err != nil {
		return fmt.Errorf("invalid end_time format: %w", err)
	}

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
		startTime, // Gunakan time.Time
		endTime,   // Gunakan time.Time
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
		var startTime, endTime time.Time

		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.DoctorID,
			&c.PetType,
			&c.PetName,
			&c.PetAge,
			&c.DiseaseDescription,
			&c.ConsultationDate,
			&startTime, // Baca sebagai time.Time
			&endTime,   // Baca sebagai time.Time
			&c.PaymentProof,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Konversi ke format response
		loc, _ := time.LoadLocation("Asia/Jakarta")
		c.StartTimeUTC = startTime.UTC().Format(time.RFC3339)
		c.EndTimeUTC = endTime.UTC().Format(time.RFC3339)
		c.StartTimeWIB = startTime.In(loc).Format("2006-01-02 15:04:05")
		c.EndTimeWIB = endTime.In(loc).Format("2006-01-02 15:04:05")

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

func (r *consultationRepo) GetDoctorDetails(ctx context.Context, doctorID int) (DoctorSchedule, error) {
	var schedule DoctorSchedule
	var workingDays, workingHours []byte

	query := `SELECT 
        gmeet_link, 
        price_per_hour,
        working_days,
        working_hours 
    FROM doctors WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, doctorID).Scan(
		&schedule.GMeetLink,
		&schedule.Price,
		&workingDays,
		&workingHours,
	)

	if err != nil {
		return DoctorSchedule{}, err
	}

	// Parse working days
	if err := json.Unmarshal(workingDays, &schedule.WorkingDays); err != nil {
		return DoctorSchedule{}, fmt.Errorf("invalid working days format: %w", err)
	}

	// Parse working hours
	if err := json.Unmarshal(workingHours, &schedule.WorkingHours); err != nil {
		return DoctorSchedule{}, fmt.Errorf("invalid working hours format: %w", err)
	}

	return schedule, nil
}

// Validasi ganda untuk dokter dan user
func (r *consultationRepo) CheckAvailability(ctx context.Context, doctorID, userID int, start, end time.Time) (bool, error) {
	query := `
    	SELECT NOT EXISTS(
        SELECT 1 FROM consultations 
        WHERE (doctor_id = $1 OR user_id = $2)
        AND (start_time, end_time) OVERLAPS ($3, $4)
	)`

	var available bool
	err := r.db.QueryRowContext(ctx, query, doctorID, userID, start, end).Scan(&available)
	return available, err
}

func (r *consultationRepo) DoctorExists(ctx context.Context, doctorID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM doctors WHERE id = $1)`
	err := r.db.QueryRowContext(ctx, query, doctorID).Scan(&exists)
	return exists, err
}
