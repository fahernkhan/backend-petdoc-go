package admin

import (
	"context"
	"time"
)

type AdminService interface {
	CreateAdmin(ctx context.Context, req CreateAdminRequest) (AdminResponse, error)
}

type adminService struct {
	repo AdminRepository
}

func NewAdminService(repo AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) CreateAdmin(ctx context.Context, req CreateAdminRequest) (AdminResponse, error) {
	currentRole, err := s.repo.GetUserRole(ctx, req.UserID)
	if err != nil {
		return AdminResponse{}, err
	}

	if currentRole != "user" {
		return AdminResponse{}, ErrInvalidRoleTransition
	}

	if err := s.repo.UpdateUserRole(ctx, req.UserID, "admin"); err != nil {
		return AdminResponse{}, err
	}

	return AdminResponse{
		Message:      "User berhasil di-promote menjadi admin",
		UserID:       req.UserID,
		PreviousRole: currentRole,
		NewRole:      "admin",
		PromotedAt:   time.Now(),
	}, nil
}
