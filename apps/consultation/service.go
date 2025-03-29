package consultation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"petdoc/internal/infrastructure/cloudinary"
	"strings"
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
	// 1. Buka file
	file, err := req.PaymentProof.Open()
	if err != nil {
		return ConsultationResponse{}, err
	}
	defer file.Close()

	// 2. Baca file ke byte (untuk validasi MIME type)
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return ConsultationResponse{}, err
	}

	// 3. Validasi MIME Type di sini <--- TEMPATKAN DI SINI
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	mimeType := http.DetectContentType(fileBytes)
	if !allowedTypes[mimeType] {
		return ConsultationResponse{}, errors.New("hanya menerima file JPEG/PNG")
	}

	// 4. Konversi ke data URI
	// base64String := base64.StdEncoding.EncodeToString(fileBytes)
	// dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64String)
	// 4. Buat reader dari byte yang sudah dibaca
	fileReader := bytes.NewReader(fileBytes) // ✅ Reset otomatis ke awal

	// 5. Upload ke Cloudinary
	paymentURL, err := s.cloudinary.Upload(ctx, cloudinary.UploadParams{
		File:     fileReader,
		Folder:   "payment_proofs",
		PublicID: fmt.Sprintf("payment_%d_%d", req.UserID, time.Now().Unix()),
	})
	if err != nil {
		// Tambahkan logging detail
		s.logger.Error("Gagal upload Cloudinary",
			"error", err,
			"user_id", req.UserID,
			"file_size", len(fileBytes),
		)
		return ConsultationResponse{}, fmt.Errorf("gagal upload bukti bayar: %w", err)
	}
	// 2. Parse waktu
	startTimeUTC, endTimeUTC, err := s.parseTimes(req.ConsultationDate, req.StartTime, req.EndTime)
	if err != nil {
		return ConsultationResponse{}, err
	}
	//cek waktu
	fmt.Println("Server Time Now:", time.Now().Format(time.RFC3339))
	fmt.Println("Start Time Parsed:", startTimeUTC)
	fmt.Println("End Time Parsed:", endTimeUTC)
	fmt.Println("Current Time:", time.Now())
	//
	// 3. Validasi waktu dasar
	if err := s.validateBasicTimes(startTimeUTC, endTimeUTC); err != nil {
		return ConsultationResponse{}, err
	}

	// // Cek apakah dokter exists
	// // 4. Cek eksistensi dokter
	// exists, err := s.repo.DoctorExists(ctx, req.DoctorID)
	// if err != nil || !exists {
	// 	return ConsultationResponse{}, ErrDoctorNotFound
	// }

	// Konversi ke WIB untuk response
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startTimeWIB := startTimeUTC.In(loc).Format("2006-01-02 15:04:05")
	endTimeWIB := endTimeUTC.In(loc).Format("2006-01-02 15:04:05")

	// 4. Cek eksistensi dokter
	exists, err := s.repo.DoctorExists(ctx, req.DoctorID)
	if err != nil || !exists {
		return ConsultationResponse{}, ErrDoctorNotFound
	}

	// 5. Ambil jadwal dokter
	doctorSchedule, err := s.repo.GetDoctorDetails(ctx, req.DoctorID)
	if err != nil {
		return ConsultationResponse{}, ErrDoctorNotFound
	}

	// 6. Validasi jadwal dokter
	if err := s.validateDoctorSchedule(startTimeUTC, endTimeUTC, doctorSchedule); err != nil {
		return ConsultationResponse{}, err
	}

	// 7. Cek ketersediaan dokter
	available, err := s.repo.CheckDoctorAvailability(ctx, req.DoctorID, startTimeUTC, endTimeUTC)
	if err != nil || !available {
		return ConsultationResponse{}, ErrDoctorNotAvailable
	}

	// 8. Cek konflik ganda (dokter dan user)
	available, err = s.repo.CheckAvailability(ctx, req.DoctorID, req.UserID, startTimeUTC, endTimeUTC)
	if err != nil || !available {
		return ConsultationResponse{}, ErrDuplicateBooking
	}

	// ... [bagian pembuatan objek konsultasi]

	// 6. Buat objek konsultasi
	// Konversi ke format response
	consultationDateStr := req.ConsultationDate // Sudah dalam format YYYY-MM-DD
	startTimeUTCStr := startTimeUTC.Format(time.RFC3339)
	endTimeUTCStr := endTimeUTC.Format(time.RFC3339)
	cons := ConsultationResponse{
		UserID:             req.UserID,
		DoctorID:           req.DoctorID,
		PetType:            req.PetType,
		PetName:            req.PetName,
		PetAge:             req.PetAge,
		DiseaseDescription: req.DiseaseDescription,
		ConsultationDate:   consultationDateStr,
		StartTimeUTC:       startTimeUTCStr,
		EndTimeUTC:         endTimeUTCStr,
		StartTimeWIB:       startTimeWIB,
		EndTimeWIB:         endTimeWIB,
		PaymentProof:       paymentURL,
	}

	// 7. Simpan ke database
	if err := s.repo.CreateConsultation(ctx, &cons); err != nil {
		s.logger.Error("Gagal menyimpan konsultasi", "error", err)
		return ConsultationResponse{}, errors.New("gagal menyimpan konsultasi")
	}

	cons.MeetLink = doctorSchedule.GMeetLink // Pakai GMeetLink dari doctorSchedule
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
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Format yang diharapkan
	dateLayout := "2006-01-02"
	timeLayout := "15:04"

	// Parse tanggal
	consultationDate, err := time.ParseInLocation(dateLayout, date, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid consultation date: %v", err)
	}

	// Parse waktu mulai
	startTime, err := time.ParseInLocation(timeLayout, start, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start time: %v", err)
	}

	// Parse waktu selesai
	endTime, err := time.ParseInLocation(timeLayout, end, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end time: %v", err)
	}

	// Gabungkan tanggal dan waktu
	startCombined := time.Date(
		consultationDate.Year(),
		consultationDate.Month(),
		consultationDate.Day(),
		startTime.Hour(),
		startTime.Minute(),
		0, // Second
		0, // Nanosecond
		loc,
	).UTC()

	endCombined := time.Date(
		consultationDate.Year(),
		consultationDate.Month(),
		consultationDate.Day(),
		endTime.Hour(),
		endTime.Minute(),
		0,
		0,
		loc,
	).UTC()

	return startCombined, endCombined, nil
}

func (s *consultationService) validateTimes(start, end time.Time) error {
	now := time.Now().UTC()
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// 1. Tidak boleh di masa lalu
	if start.Before(now) {
		return ErrConsultationPastDate
	}

	// 2. Durasi minimal 30 menit
	if end.Sub(start) < 30*time.Minute {
		return errors.New("durasi minimal 30 menit")
	}

	// 3. End time harus setelah start time
	if !end.After(start) {
		return errors.New("end_time harus setelah start_time")
	}

	// 4. Validasi jam kerja 08:00-20:00 WIB
	startWIB := start.In(loc)
	endWIB := end.In(loc)

	if startWIB.Hour() < 8 || startWIB.Hour() >= 20 {
		return errors.New("jam mulai harus antara 08:00 - 20:00 WIB")
	}

	if endWIB.Hour() < 8 || endWIB.Hour() > 20 {
		return errors.New("jam selesai harus antara 08:00 - 20:00 WIB")
	}

	return nil
}

// Tambahkan fungsi helper di bawah service.go
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func (s *consultationService) validateBasicTimes(start, end time.Time) error {
	now := time.Now().UTC()

	// 1. Tidak boleh di masa lalu
	if start.Before(now) {
		return ErrConsultationPastDate
	}

	// 2. Durasi minimal 30 menit
	if end.Sub(start) < 30*time.Minute {
		return errors.New("durasi minimal 30 menit")
	}

	// 3. End time harus setelah start time
	if !end.After(start) {
		return errors.New("end_time harus setelah start_time")
	}

	return nil
}

func (s *consultationService) validateDoctorSchedule(start, end time.Time, schedule DoctorSchedule) error {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startWIB := start.In(loc)
	endWIB := end.In(loc)

	// Validasi hari kerja
	dayOfWeek := startWIB.Weekday().String()
	if !contains(schedule.WorkingDays, dayOfWeek) {
		return fmt.Errorf("dokter tidak tersedia pada %s", dayOfWeek)
	}

	// Validasi format jam kerja dokter
	if _, err := time.Parse("15:04", schedule.WorkingHours.Start); err != nil {
		return fmt.Errorf("format jam mulai dokter tidak valid: %w", err)
	}

	if _, err := time.Parse("15:04", schedule.WorkingHours.End); err != nil {
		return fmt.Errorf("format jam selesai dokter tidak valid: %w", err)
	}

	// Ekstrak jam konsultasi
	consultStart := fmt.Sprintf("%02d:%02d", startWIB.Hour(), startWIB.Minute())
	consultEnd := fmt.Sprintf("%02d:%02d", endWIB.Hour(), endWIB.Minute())

	// Validasi jam mulai
	if consultStart < schedule.WorkingHours.Start {
		return fmt.Errorf("jam mulai harus setelah %s", schedule.WorkingHours.Start)
	}

	// Validasi jam selesai
	if consultEnd > schedule.WorkingHours.End {
		return fmt.Errorf("jam selesai harus sebelum %s", schedule.WorkingHours.End)
	}

	return nil
}
