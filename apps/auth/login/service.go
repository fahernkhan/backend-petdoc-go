package login

import (
	"log/slog"
	"petdoc/internal/infrastructure/utils/encryption"
	"petdoc/internal/infrastructure/utils/jwt"
	"time"

	"golang.org/x/net/context"
)

// Service adalah interface untuk layanan login.
// Interface ini mendefinisikan satu metode utama, yaitu Login.
type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

// Struktur `service` adalah implementasi dari `Service`.
// Ini berisi dependensi yang dibutuhkan untuk menangani proses login.
type service struct {
	repo     Repository    // Repository untuk mengambil data pengguna dari database.
	jwt      jwt.JWT       // Komponen untuk menangani pembuatan dan validasi token JWT.
	tokenExp time.Duration // Durasi masa berlaku token JWT.
}

// NewService adalah fungsi konstruktor yang membuat instance baru dari `service`.
// Fungsi ini menerima repository, handler JWT, dan durasi token sebagai parameter.
func NewService(repo Repository, jwt jwt.JWT, tokenExp time.Duration) Service {
	return &service{
		repo:     repo,
		jwt:      jwt,
		tokenExp: tokenExp,
	}
}
func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	logger := slog.With(
		"module", "login",
		"email", req.Email,
	)

	// 1. Cari user berdasarkan email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Warn("User not found", "error", err)
		return nil, ErrInvalidCredentials
	}

	// 2. Validasi password

	if err := encryption.ValidatePassword(user.Password, req.Password); err != nil {
		logger.Warn("Invalid password", "error", err)
		return nil, ErrInvalidCredentials
	}

	// 3. Generate JWT token
	claims := jwt.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}

	token, expiry, err := s.jwt.GenerateToken(claims, s.tokenExp)
	if err != nil {
		logger.Error("Failed to generate token", "error", err)
		return nil, err
	}

	// 4. Format response
	response := &LoginResponse{
		AccessToken: token,
		ExpiresAt:   expiry,
		User: UserResponse{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
			Username: user.Username,
			Role:     user.Role,
		},
	}

	logger.Info("Login successful", "user_id", user.ID)
	return response, nil
}
