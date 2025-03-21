package register

import (
	"context"
	"log/slog"
	"petdoc/internal/infrastructure/utils/encryption"
	"time"
)

// Timeout untuk operasi database (misalnya 5 detik)
const dbTimeout = 5 * time.Second

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (User, error)
}

type service struct {
	repo Repository
}

// NewService membuat instance service baru
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (User, error) {
	// Buat context dengan timeout
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	logger := slog.With("module", "register-service")

	// 1. Validasi email unik
	emailExists, err := s.repo.CheckEmailExists(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to check email existence", "error", err)
		return User{}, err
	}
	if emailExists {
		logger.Warn("Email already exists", "email", req.Email)
		return User{}, ErrEmailExists
	}

	// 2. Validasi username unik
	usernameExists, err := s.repo.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		logger.Error("Failed to check username existence", "error", err)
		return User{}, err
	}
	if usernameExists {
		logger.Warn("Username already exists", "username", req.Username)
		return User{}, ErrUsernameExists
	}

	// 3. Hash password
	hashedPassword, err := encryption.GenerateFromPassword(req.Password)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return User{}, err
	}

	// 4. Buat user
	user := User{
		FullName:    req.FullName,
		Email:       req.Email,
		Password:    hashedPassword,
		Gender:      req.Gender,
		Username:    req.Username,
		DateOfBirth: req.DateOfBirth,
		Role:        "user", // Default role
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 5. Simpan ke database (gunakan pointer!)
	if err := s.repo.CreateUser(ctx, &user); err != nil { // ✅ Kirim pointer
		logger.Error("Failed to create user", "error", err)
		return User{}, err
	}

	logger.Info("User registered successfully", "user_id", user.ID, "email", user.Email)

	return user, nil
}
