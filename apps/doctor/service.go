package doctor

import (
	"context"
	"log/slog"
	"time"
)

type Service interface {
	CreateDoctor(ctx context.Context, req DoctorRequest) (DoctorResponse, error)
	GetDoctor(ctx context.Context, id int) (DoctorResponse, error)
	ListDoctors(ctx context.Context, pagination Pagination) ([]DoctorResponse, int, error)
	UpdateDoctor(ctx context.Context, id int, req DoctorRequest) (DoctorResponse, error)
	DeleteDoctor(ctx context.Context, id int) error
}

type service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository) Service {
	return &service{
		repo:   repo,
		logger: slog.With("module", "doctor_service"),
	}
}

func (s *service) CreateDoctor(ctx context.Context, req DoctorRequest) (DoctorResponse, error) {
	if err := validateWorkingHours(req); err != nil {
		return DoctorResponse{}, err
	}

	// Validasi hari kerja(jika diperlukan)
	// if err := validateWorkingDays(req.WorkingDays); err != nil {
	// 	return DoctorResponse{}, err
	// }

	id, err := s.repo.Create(ctx, &req)
	if err != nil {
		s.logger.Error("Failed to create doctor", "error", err)
		return DoctorResponse{}, err
	}

	return s.GetDoctor(ctx, id)
}

func validateWorkingHours(req DoctorRequest) error {
	start, err := time.Parse("15:04", req.WorkingHoursStart)
	if err != nil {
		return ErrInvalidWorkingHours
	}

	end, err := time.Parse("15:04", req.WorkingHoursEnd)
	if err != nil {
		return ErrInvalidWorkingHours
	}

	if !end.After(start) {
		return ErrInvalidWorkingHours
	}

	return nil
}

// Implementasi method lainnya
// GetDoctor implementation
func (s *service) GetDoctor(ctx context.Context, id int) (DoctorResponse, error) {
	if id <= 0 {
		return DoctorResponse{}, ErrInvalidDoctorData
	}

	doctor, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get doctor", "id", id, "error", err)
		return DoctorResponse{}, err
	}

	return doctor, nil
}

// ListDoctors implementation
func (s *service) ListDoctors(ctx context.Context, pagination Pagination) ([]DoctorResponse, int, error) {
	if pagination.Page < 1 || pagination.Limit < 1 {
		return nil, 0, ErrInvalidDoctorData
	}

	doctors, total, err := s.repo.GetAll(ctx, pagination.Page, pagination.Limit)
	if err != nil {
		s.logger.Error("Failed to list doctors", "error", err)
		return nil, 0, err
	}

	return doctors, total, nil
}

// UpdateDoctor implementation
func (s *service) UpdateDoctor(ctx context.Context, id int, req DoctorRequest) (DoctorResponse, error) {
	// Validate existing doctor
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Doctor not found", "id", id, "error", err)
		return DoctorResponse{}, err
	}

	// (Opsional) Bisa tambahkan validasi data lama sebelum update
	s.logger.Info("Updating doctor", "id", id, "oldData", existing)

	// Validate working hours
	if err := validateWorkingHours(req); err != nil {
		return DoctorResponse{}, err
	}

	// Update doctor
	if err := s.repo.Update(ctx, id, &req); err != nil {
		s.logger.Error("Failed to update doctor", "id", id, "error", err)
		return DoctorResponse{}, err
	}

	// Get updated data
	updatedDoctor, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve updated doctor data", "id", id, "error", err)
		return DoctorResponse{}, err
	}

	return updatedDoctor, nil
}

// DeleteDoctor implementation
func (s *service) DeleteDoctor(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidDoctorData
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete doctor", "id", id, "error", err)
		return err
	}

	return nil
}

// Tambahkan validasi tambahan
// func validateWorkingDays(days []string) error {
// 	validDays := map[string]bool{
// 		"Senin": true, "Selasa": true, "Rabu": true,
// 		"Kamis": true, "Jumat": true, "Sabtu": true, "Minggu": true,
// 	}

// 	for _, day := range days {
// 		if !validDays[day] {
// 			return ErrInvalidWorkingDays
// 		}
// 	}
// 	return nil
// }
