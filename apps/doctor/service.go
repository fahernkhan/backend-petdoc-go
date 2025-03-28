package doctor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return DoctorResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Validasi target user
	targetUserRole, err := s.repo.GetUserRole(ctx, tx, req.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DoctorResponse{}, ErrUserNotFound
		}
		return DoctorResponse{}, fmt.Errorf("failed to get user role: %w", err)
	}

	if targetUserRole != "user" {
		return DoctorResponse{}, ErrUserNotAllowed
	}

	// Update role
	if err := s.repo.UpdateUserRole(ctx, tx, req.UserID, "doctor"); err != nil {
		return DoctorResponse{}, fmt.Errorf("failed to update role: %w", err)
	}

	// Validasi input
	if err := validateWorkingHours(req); err != nil {
		return DoctorResponse{}, fmt.Errorf("invalid working hours: %w", err)
	}

	if err := validateWorkingDays(req.WorkingDays); err != nil {
		return DoctorResponse{}, fmt.Errorf("invalid working days: %w", err)
	}

	// Create doctor
	id, err := s.repo.Create(ctx, tx, &req)
	if err != nil {
		if errors.Is(err, ErrDuplicateEntry) {
			return DoctorResponse{}, fmt.Errorf("doctor already exists: %w", err)
		}
		return DoctorResponse{}, fmt.Errorf("failed to create doctor: %w", err)
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		return DoctorResponse{}, fmt.Errorf("commit failed: %w", err)
	}

	// Get data dengan transaction baru
	newDoctor, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return DoctorResponse{}, fmt.Errorf("failed to fetch new doctor: %w", err)
	}

	return newDoctor, nil
}

// validateWorkingHours
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

// validateWorkingDays
func validateWorkingDays(days []string) error {
	validDays := map[string]bool{
		"Senin": true, "Selasa": true, "Rabu": true,
		"Kamis": true, "Jumat": true, "Sabtu": true, "Minggu": true,
	}

	for _, day := range days {
		if !validDays[day] {
			return ErrInvalidWorkingDays
		}
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
// service.go
func (s *service) DeleteDoctor(ctx context.Context, id int) error {
	// Mulai transaksi
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Dapatkan data dokter untuk ambil UserID
	// 1. Ambil data dokter
	doctor, err := s.repo.GetByIDWithTx(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("doctor not found: %w", err)
	}

	// Dapatkan role user saat ini
	// 2. Cek role user
	currentRole, err := s.repo.GetUserRole(ctx, tx, doctor.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User sudah dihapus, lanjut hapus dokter tanpa ubah role
			s.logger.Warn("user not found, proceeding with doctor deletion", "user_id", doctor.UserID)
		} else {
			return fmt.Errorf("failed to get user role: %w", err)
		}
	} else {
		// Ubah role kembali ke 'user' hanya jika saat ini 'doctor'
		// 3. Revert role jika masih 'doctor'
		if currentRole == "doctor" {
			if err := s.repo.UpdateUserRole(ctx, tx, doctor.UserID, "user"); err != nil {
				return fmt.Errorf("failed to revert user role: %w", err)
			}
		}
	}

	// Hapus dokter
	// 4. Hapus dokter
	if err := s.repo.DeleteWithTx(ctx, tx, id); err != nil {
		return fmt.Errorf("failed to delete doctor: %w", err)
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
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
