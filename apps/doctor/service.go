package doctor

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"petdoc/internal/infrastructure/cloudinary"
	"strings"
	"time"
)

type Service interface {
	CreateDoctor(ctx context.Context, req CreateRequest) (Response, error)
	GetDoctor(ctx context.Context, id int) (Response, error)
	ListDoctors(ctx context.Context, pagination Pagination, search string) ([]Response, int, error)
	UpdateDoctor(ctx context.Context, id int, req UpdateRequest) (Response, error)
	DeleteDoctor(ctx context.Context, id int) error
}

type service struct {
	repo       Repository
	cloudinary cloudinary.Service
	logger     *slog.Logger
}

func NewService(repo Repository, cloudinary cloudinary.Service, logger *slog.Logger) Service {
	return &service{
		repo:       repo,
		cloudinary: cloudinary,
		logger:     logger,
	}
}

func convertToServiceResponse(dr DoctorResponse) Response {

	return Response{
		ID:                dr.ID,
		UserID:            dr.UserID,
		FullName:          dr.FullName,
		LastEducation:     dr.LastEducation,
		SpecialistAt:      dr.SpecialistAt,
		ProfileImage:      dr.ProfileImage,
		BirthDate:         dr.BirthDate,
		HospitalName:      dr.HospitalName,
		YearsOfExperience: dr.YearsOfExperience,
		PricePerHour:      dr.PricePerHour,
		GmeetLink:         dr.GmeetLink,
		WorkingDays:       dr.WorkingDays,
		WorkingHours: struct {
			Start string `json:"start"`
			End   string `json:"end"`
		}{
			Start: dr.WorkingHours.Start,
			End:   dr.WorkingHours.End,
		},
		CreatedAt: dr.CreatedAt,
		UpdatedAt: dr.UpdatedAt,
	}
}

func (s *service) CreateDoctor(ctx context.Context, req CreateRequest) (Response, error) {
	//imageURL, err := s.uploadImage(ctx, req.ProfileImage)
	//if err != nil {
	//	return Response{}, err
	//}
	//
	//tx, err := s.repo.BeginTx(ctx)
	//if err != nil {
	//	return Response{}, fmt.Errorf("failed to begin transaction: %w", err)
	//}
	//defer tx.Rollback()
	// Sebelum upload gambar, mulai transaksi
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return Response{}, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Upload gambar diluar transaksi
	imageURL, err := s.uploadImage(ctx, req.ProfileImage)
	if err != nil {
		return Response{}, fmt.Errorf("upload image failed: %w", err)
	}

	// Parse working_days dari form-data (string JSON ke []string)
	var workingDays []string
	if err := json.Unmarshal([]byte(req.WorkingDays), &workingDays); err != nil {
		return Response{}, fmt.Errorf("invalid working_days format: %w", err)
	}

	// Validasi hari
	if err := validateWorkingDays(workingDays); err != nil {
		return Response{}, fmt.Errorf("invalid working days: %w", err)
	}
	// Transaksi hanya untuk operasi database
	defer tx.Rollback()
	repoReq := DoctorRequest{
		UserID:            req.UserID,
		FullName:          req.FullName,
		LastEducation:     req.LastEducation,
		SpecialistAt:      req.SpecialistAt,
		ProfileImage:      imageURL,
		BirthDate:         req.BirthDate,
		HospitalName:      req.HospitalName,
		YearsOfExperience: req.YearsOfExperience,
		PricePerHour:      req.PricePerHour,
		GmeetLink:         req.GmeetLink,
		WorkingDays:       workingDays,
		WorkingHoursStart: req.WorkingHoursStart,
		WorkingHoursEnd:   req.WorkingHoursEnd,
	}

	targetUserRole, err := s.repo.GetUserRole(ctx, tx, repoReq.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Response{}, ErrUserNotFound
		}
		return Response{}, fmt.Errorf("failed to get user role: %w", err)
	}

	if targetUserRole != "user" {
		return Response{}, ErrUserNotAllowed
	}

	if err := s.repo.UpdateUserRole(ctx, tx, repoReq.UserID, "doctor"); err != nil {
		return Response{}, fmt.Errorf("failed to update role: %w", err)
	}

	if err := validateWorkingHours(repoReq); err != nil {
		return Response{}, fmt.Errorf("invalid working hours: %w", err)
	}

	if err := validateWorkingDays(repoReq.WorkingDays); err != nil {
		return Response{}, fmt.Errorf("invalid working days: %w", err)
	}

	id, err := s.repo.Create(ctx, tx, &repoReq)
	if err != nil {
		if errors.Is(err, ErrDuplicateEntry) {
			return Response{}, fmt.Errorf("doctor already exists: %w", err)
		}
		return Response{}, fmt.Errorf("failed to create doctor: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Response{}, fmt.Errorf("commit failed: %w", err)
	}

	newDoctorResp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf("failed to fetch new doctor: %w", err)
	}

	return convertToServiceResponse(newDoctorResp), nil
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
		"Monday": true, "Tuesday": true, "Wednesday": true,
		"Thursday": true, "Friday": true, "Saturday": true, "Sunday": true,
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
func (s *service) GetDoctor(ctx context.Context, id int) (Response, error) {
	if id <= 0 {
		return Response{}, ErrInvalidDoctorData
	}

	doctorResp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get doctor", "id", id, "error", err)
		return Response{}, err
	}

	// Convert DoctorResponse ke Response
	return convertToServiceResponse(doctorResp), nil
}

// ListDoctors implementation
func (s *service) ListDoctors(ctx context.Context, pagination Pagination, search string) ([]Response, int, error) {
	if pagination.Page < 1 || pagination.Limit < 1 {
		return nil, 0, ErrInvalidDoctorData
	}

	var (
		doctorResponses []DoctorResponse
		total           int
		err             error
	)

	if search != "" {
		doctorResponses, total, err = s.repo.Search(ctx, search, pagination.Page, pagination.Limit)
	} else {
		doctorResponses, total, err = s.repo.GetAll(ctx, pagination.Page, pagination.Limit)
	}

	if err != nil {
		s.logger.Error("Failed to list doctors", "error", err)
		return nil, 0, err
	}

	responses := make([]Response, len(doctorResponses))
	for i, dr := range doctorResponses {
		responses[i] = convertToServiceResponse(dr)
	}

	return responses, total, nil
}

//func (s *service) ListDoctors(ctx context.Context, pagination Pagination) ([]Response, int, error) {
//	if pagination.Page < 1 || pagination.Limit < 1 {
//		return nil, 0, ErrInvalidDoctorData
//	}
//
//	doctorResponses, total, err := s.repo.GetAll(ctx, pagination.Page, pagination.Limit)
//	if err != nil {
//		s.logger.Error("Failed to list doctors", "error", err)
//		return nil, 0, err
//	}
//
//	// Convert semua DoctorResponse ke Response
//	responses := make([]Response, len(doctorResponses))
//	for i, dr := range doctorResponses {
//		responses[i] = convertToServiceResponse(dr)
//	}
//
//	return responses, total, nil
//}

// UpdateDoctor implementation
func (s *service) UpdateDoctor(ctx context.Context, id int, req UpdateRequest) (Response, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Doctor not found", "id", id, "error", err)
		return Response{}, err
	}

	var imageURL string
	if req.ProfileImage != nil {
		newImageURL, err := s.uploadImage(ctx, req.ProfileImage)
		if err != nil {
			return Response{}, err
		}
		imageURL = newImageURL

		if existing.ProfileImage != "" {
			publicID := extractPublicID(existing.ProfileImage)
			if publicID != "" {
				if err := s.cloudinary.Delete(ctx, publicID); err != nil {
					s.logger.Error("Gagal hapus gambar lama", "publicID", publicID, "error", err)
				}
			}
		}
	} else {
		imageURL = existing.ProfileImage
	}

	// Parse working_days dari string ke slice
	var workingDays []string
	if err := json.Unmarshal([]byte(req.WorkingDays), &workingDays); err != nil {
		return Response{}, fmt.Errorf("invalid working_days format: %w", err)
	}

	// Validasi hari
	if err := validateWorkingDays(workingDays); err != nil {
		return Response{}, fmt.Errorf("invalid working days: %w", err)
	}

	// Mapping ke repository request
	updateData := DoctorRequest{
		FullName:          req.FullName,
		LastEducation:     req.LastEducation,
		SpecialistAt:      req.SpecialistAt,
		ProfileImage:      imageURL,
		BirthDate:         req.BirthDate,
		HospitalName:      req.HospitalName,
		YearsOfExperience: req.YearsOfExperience,
		PricePerHour:      req.PricePerHour,
		GmeetLink:         req.GmeetLink,
		WorkingDays:       workingDays, // Gunakan slice hasil parsing
		WorkingHoursStart: req.WorkingHoursStart,
		WorkingHoursEnd:   req.WorkingHoursEnd,
	}

	if err := s.repo.Update(ctx, id, &updateData); err != nil {
		s.logger.Error("Failed to update doctor", "id", id, "error", err)
		return Response{}, err
	}

	updatedDoctorResp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to retrieve updated doctor data", "id", id, "error", err)
		return Response{}, err
	}
	return convertToServiceResponse(updatedDoctorResp), nil
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

func (s *service) uploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// Buka file
	src, err := file.Open()
	if err != nil {
		s.logger.Error("Gagal membuka file", "error", err)
		return "", fmt.Errorf("%w: gagal membuka file", ErrImageUploadFailed)
	}
	defer src.Close()

	// Validasi ukuran file
	if file.Size > 5*1024*1024 { // 5MB
		return "", ErrFileTooLarge
	}
	// Baca seluruh file ke buffer
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		s.logger.Error("Gagal membaca file", "error", err)
		return "", fmt.Errorf("%w: gagal membaca file", ErrImageUploadFailed)
	}

	// Validasi tipe file dari buffer
	mimeType := http.DetectContentType(fileBytes)
	if !isValidImageType(mimeType) {
		s.logger.Error("Format gambar tidak valid", "mime_type", mimeType)
		return "", ErrInvalidImageFormat
	}

	// Upload dari buffer (bukan dari reader yang sudah habis)
	publicID := fmt.Sprintf("doctor_%d", time.Now().UnixNano())
	url, err := s.cloudinary.Upload(ctx, cloudinary.UploadParams{
		File:     bytes.NewReader(fileBytes), // <-- Gunakan buffer
		Folder:   "doctors",
		PublicID: publicID,
	})

	if err != nil {
		s.logger.Error("Upload ke Cloudinary gagal", "error", err)
		return "", fmt.Errorf("%w: %v", ErrImageUploadFailed, err)
	}
	s.logger.Debug("Upload gambar berhasil", "url", url)
	return url, nil
}

//func (s *service) uploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
//	src, err := file.Open()
//	if err != nil {
//		return "", ErrImageUploadFailed
//	}
//	defer src.Close()
//
//	fileBytes, err := io.ReadAll(src)
//	if err != nil {
//		return "", ErrImageUploadFailed
//	}
//
//	// Validate image type
//	mimeType := http.DetectContentType(fileBytes)
//	if !isValidImageType(mimeType) {
//		return "", ErrInvalidImageFormat
//	}
//
//	// Stream langsung ke Cloudinary
//	publicID := fmt.Sprintf("doctor_%d", time.Now().UnixNano())
//	//imageReader := bytes.NewReader(fileBytes)
//
//	url, err := s.cloudinary.Upload(ctx, cloudinary.UploadParams{
//		//File:     imageReader,
//		File:     src,
//		Folder:   "doctors",
//		PublicID: publicID,
//	})
//
//	if err != nil {
//		s.logger.Error("Image upload failed", "error", err)
//		return "", ErrImageUploadFailed
//	}
//
//	return url, nil
//}

func isValidImageType(mimeType string) bool {
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	return allowed[mimeType]
}

func extractPublicID(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) < 8 {
		return ""
	}
	return strings.Join(parts[7:len(parts)-1], "/") + "/" + strings.Split(parts[len(parts)-1], ".")[0]
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
