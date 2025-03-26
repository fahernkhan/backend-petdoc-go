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

// mengimplementasikan UserService, yang merupakan service layer dalam arsitektur aplikasi
type userService struct {
	repo UserRepository
}

// NewUserService membuat instance baru dari UserService
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

// GetPaginatedUsers mengambil daftar pengguna dengan fitur pagination
func (s *userService) GetPaginatedUsers(ctx context.Context, req PaginationRequest) (*PaginatedResponse, error) {
	// Validasi parameter pagination
	if req.Page < 1 {
		return nil, &ValidationError{Field: "Page", Message: "must be at least 1"}
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		return nil, &ValidationError{Field: "pageSize", Message: "must be between 1 and 100"}
	}

	// Hitung offset untuk query database
	offset := (req.Page - 1) * req.PageSize

	// Gunakan channel untuk menjalankan query secara paralel
	dataChan := make(chan []UserResponse, 1)
	countChan := make(chan int, 1)
	errChan := make(chan error, 2)

	// Ambil data pengguna
	go func() {
		slog.Info("Fetching users from database", "offset", offset, "limit", req.PageSize)
		users, err := s.repo.GetAllUsers(ctx, offset, req.PageSize)
		errChan <- err
		dataChan <- users
	}()

	// Hitung total pengguna
	go func() {
		slog.Info("Counting total users in database")
		count, err := s.repo.CountAllUsers(ctx)
		errChan <- err
		countChan <- count
	}()

	// Tunggu semua goroutine selesai
	var users []UserResponse
	var total int
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
	users = <-dataChan
	total = <-countChan

	// Hitung metadata pagination
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	if req.Page > totalPages && totalPages > 0 {
		slog.Warn("Requested page is out of range", "requestedPage", req.Page, "totalPages", totalPages)
		return nil, ErrPageOutOfRange
	}

	fromItem := (req.Page-1)*req.PageSize + 1
	toItem := req.Page * req.PageSize
	if toItem > total {
		toItem = total
	}

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
