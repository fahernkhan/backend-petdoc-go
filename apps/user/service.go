package user

import (
	"context"
	"errors"
	"log/slog"
	"math"
)

// UserService mendefinisikan layanan untuk mendapatkan pengguna dengan pagination
type UserService interface {
	GetPaginatedUsers(ctx context.Context, req PaginationRequest) (*PaginatedResponse, error)
}

// userService mengimplementasikan UserService sebagai service layer
// yang menangani logika bisnis sebelum memanggil repository.
type userService struct {
	repo UserRepository
}

// NewUserService membuat instance baru dari UserService
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

// GetPaginatedUsers mengambil daftar pengguna dengan fitur pagination
func (s *userService) GetPaginatedUsers(ctx context.Context, req PaginationRequest) (*PaginatedResponse, error) {
	// Set default values
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// Validasi parameter pagination
	if req.Page < 1 {
		return nil, &ValidationError{Field: "Page", Message: "must be at least 1"}
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		return nil, &ValidationError{Field: "pageSize", Message: "must be between 1 and 100"}
	}

	// Hitung offset untuk query database berdasarkan page dan pageSize
	offset := (req.Page - 1) * req.PageSize

	// Gunakan channel untuk menjalankan query secara paralel agar lebih efisien
	dataChan := make(chan []UserResponse, 1) // Menampung data user dari DB
	countChan := make(chan int, 1)           // Menampung jumlah total user
	errChan := make(chan error, 2)           // Menampung error jika terjadi di goroutine

	// Goroutine untuk mengambil data pengguna dari database
	go func() {
		slog.Info("Fetching users from database", "offset", offset, "limit", req.PageSize)
		users, err := s.repo.GetAllUsers(ctx, offset, req.PageSize, req.Filter)
		errChan <- err    // Kirim error ke channel jika ada
		dataChan <- users // Kirim data ke channel jika berhasil
	}()

	// Goroutine untuk menghitung total pengguna di database
	go func() {
		slog.Info("Counting total users in database")
		count, err := s.repo.CountAllUsers(ctx, req.Filter)
		errChan <- err     // Kirim error ke channel jika ada
		countChan <- count // Kirim jumlah total pengguna
	}()

	// Variabel untuk menyimpan hasil dari goroutine
	var users []UserResponse
	var total int

	// Menunggu kedua goroutine selesai (2 iterasi karena ada 2 goroutine)
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			var dbErr *DatabaseError
			if errors.As(err, &dbErr) {
				slog.Error("Database query failed", "operation", dbErr.Operation, "error", dbErr.Err)
				return nil, ErrDatabaseQuery
			}
			return nil, err
		}
	}

	// Ambil hasil dari channel setelah memastikan tidak ada error
	users = <-dataChan
	total = <-countChan

	// Hitung total halaman berdasarkan total pengguna dan pageSize
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	if req.Page > totalPages && totalPages > 0 {
		slog.Warn("Requested page is out of range", "requestedPage", req.Page, "totalPages", totalPages)
		return nil, ErrPageOutOfRange
	}

	// Hitung item yang sedang ditampilkan pada halaman ini
	fromItem := (req.Page-1)*req.PageSize + 1
	toItem := req.Page * req.PageSize
	if toItem > total {
		toItem = total
	}

	// Buat response yang akan dikembalikan
	response := &PaginatedResponse{
		StatusCode: 200,
		Message:    "Success",
		PageNumber: req.Page,
		TotalPages: totalPages,
		FromItem:   fromItem,
		ToItem:     toItem,
		TotalItem:  total,
		Data:       users,
	}

	slog.Info("Successfully fetched paginated users", "page", req.Page, "pageSize", req.PageSize, "totalUsers", total)
	return response, nil
}
