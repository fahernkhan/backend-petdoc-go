package consultation

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"petdoc/internal/infrastructure/cloudinary"
	"time"
)

type Service interface {
	CreateConsultation(ctx context.Context, req CreateRequest) (ConsultationResponse, error)
	GetConsultations(ctx context.Context, userID, page, pageSize int) (PaginationResponse, error)
}

type consultationService struct {
	repo       Repository
	cloudinary cloudinary.Service
	logger     *slog.Logger
}

func NewService(repo Repository, cloudinary cloudinary.Service, logger *slog.Logger) Service {
	return &consultationService{
		repo:       repo,
		cloudinary: cloudinary,
		logger:     logger,
	}
}

func (s *consultationService) CreateConsultation(ctx context.Context, req CreateRequest) (ConsultationResponse, error) {
	// 1. Upload payment proof
	// Buka file
	file, err := req.PaymentProof.Open()
	if err != nil {
		return ConsultationResponse{}, err
	}
	defer file.Close()

	// Baca file ke byte
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return ConsultationResponse{}, err
	}

	// Konversi ke base64
	base64String := base64.StdEncoding.EncodeToString(fileBytes)

	// Upload ke Cloudinary
	paymentURL, err := s.cloudinary.Upload(ctx, cloudinary.UploadParams{
		File:     base64String,
		Folder:   "payment_proofs",
		PublicID: fmt.Sprintf("payment_%d_%d", req.UserID, time.Now().Unix()),
	})
	// 2. Parse waktu
	startTime, endTime, err := s.parseTimes(req.ConsultationDate, req.StartTime, req.EndTime)
	if err != nil {
		return ConsultationResponse{}, err
	}
	//cek waktu
	fmt.Println("Server Time Now:", time.Now().Format(time.RFC3339))
	fmt.Println("Start Time Parsed:", startTime)
	fmt.Println("End Time Parsed:", endTime)
	fmt.Println("Current Time:", time.Now())
	// 3. Validasi waktu
	if err := s.validateTimes(startTime, endTime); err != nil {
		return ConsultationResponse{}, err
	}

	// 4. Cek ketersediaan dokter
	available, err := s.repo.CheckDoctorAvailability(ctx, req.DoctorID, startTime, endTime)
	if err != nil || !available {
		return ConsultationResponse{}, ErrDoctorNotAvailable
	}

	// 5. Dapatkan detail dokter
	gmeetLink, _, err := s.repo.GetDoctorDetails(ctx, req.DoctorID)
	if err != nil {
		return ConsultationResponse{}, ErrDoctorNotFound
	}

	// 6. Buat objek konsultasi
	consultationDate, _ := time.Parse("2006-01-02", req.ConsultationDate)
	cons := ConsultationResponse{
		UserID:             req.UserID,
		DoctorID:           req.DoctorID,
		PetType:            req.PetType,
		PetName:            req.PetName,
		PetAge:             req.PetAge,
		DiseaseDescription: req.DiseaseDescription,
		ConsultationDate:   consultationDate,
		StartTime:          startTime,
		EndTime:            endTime,
		PaymentProof:       paymentURL,
	}

	// 7. Simpan ke database
	if err := s.repo.CreateConsultation(ctx, &cons); err != nil {
		s.logger.Error("Gagal menyimpan konsultasi", "error", err)
		return ConsultationResponse{}, errors.New("gagal menyimpan konsultasi")
	}

	cons.MeetLink = gmeetLink
	return cons, nil
}

func (s *consultationService) GetConsultations(ctx context.Context, userID, page, pageSize int) (PaginationResponse, error) {
	consults, total, err := s.repo.GetConsultations(ctx, userID, page, pageSize)
	if err != nil {
		return PaginationResponse{}, err
	}

	totalPages := int(total)/pageSize + 1
	return PaginationResponse{
		Data:       consults,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

// Helper functions
func (s *consultationService) parseTimes(date, start, end string) (time.Time, time.Time, error) {
	// Gabungkan tanggal dengan waktu
	startFull := fmt.Sprintf("%sT%s", date, start) // "2024-03-28T14:00:00+07:00"
	endFull := fmt.Sprintf("%sT%s", date, end)     // "2024-03-28T15:00:00+07:00"

	layout := "2006-01-02T15:04:05Z07:00" // Format dengan timezone

	startTime, err := time.Parse(layout, startFull)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start time: %v", err)
	}

	endTime, err := time.Parse(layout, endFull)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end time: %v", err)
	}

	return startTime.UTC(), endTime.UTC(), nil
}
func (s *consultationService) validateTimes(start, end time.Time) error {
	now := time.Now().UTC() // Gunakan UTC untuk konsistensi

	if start.Before(now) {
		return ErrConsultationPastDate
	}

	if end.Sub(start).Minutes() < 30 {
		return errors.New("durasi minimal 30 menit")
	}

	return nil
}
