package clinic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"petdoc/internal/infrastructure/cloudinary"
	"time"
)

type Service interface {
	CreateClinic(ctx context.Context, req CreateRequest, userID int) (ClinicResponse, error)
	GetClinic(ctx context.Context, id int) (ClinicResponse, error)
	GetAllClinics(ctx context.Context, query string, page, pageSize int) (PaginationResponse, error)
	UpdateClinic(ctx context.Context, req UpdateRequest, id, userID int) (ClinicResponse, error)
	DeleteClinic(ctx context.Context, id, userID int) error
	SearchClinics(ctx context.Context, query string, page, pageSize int) (PaginationResponse, error)
}

type clinicService struct {
	repo       Repository
	cloudinary cloudinary.Service
	logger     *slog.Logger
}

func NewService(repo Repository, cloudinary cloudinary.Service, logger *slog.Logger) Service {
	return &clinicService{
		repo:       repo,
		cloudinary: cloudinary,
		logger:     logger,
	}
}

func (s *clinicService) CreateClinic(ctx context.Context, req CreateRequest, userID int) (ClinicResponse, error) {
	imageURL, err := s.uploadImage(ctx, req.Image)
	if err != nil {
		return ClinicResponse{}, err
	}

	clinic := ClinicResponse{
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		MapLink:     req.MapLink,
		ImageURL:    imageURL,
		UserID:      userID,
	}

	if err := s.repo.Create(ctx, &clinic); err != nil {
		s.logger.Error("Failed to create clinic", "error", err)
		return ClinicResponse{}, ErrInvalidClinicData
	}

	return clinic, nil
}

func (s *clinicService) GetClinic(ctx context.Context, id int) (ClinicResponse, error) {
	clinic, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get clinic", "id", id, "error", err)
		return ClinicResponse{}, ErrClinicNotFound
	}
	return clinic, nil
}

func (s *clinicService) GetAllClinics(ctx context.Context, query string, page, pageSize int) (PaginationResponse, error) {
	if page < 1 || pageSize < 1 {
		return PaginationResponse{}, ErrPaginationInvalid
	}

	var clinics []ClinicResponse
	var total int64
	var err error

	if query != "" {
		clinics, total, err = s.repo.Search(ctx, query, page, pageSize)
	} else {
		clinics, total, err = s.repo.GetAll(ctx, page, pageSize)
	}

	if err != nil {
		s.logger.Error("Failed to get clinics", "error", err)
		return PaginationResponse{}, err
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	fromItem := (page-1)*pageSize + 1
	toItem := page * pageSize
	if toItem > int(total) {
		toItem = int(total)
	}

	return PaginationResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Message:    "Success",
		PageNumber: page,
		TotalPages: totalPages,
		FromItem:   fromItem,
		ToItem:     toItem,
		TotalItem:  total,
		Data:       clinics,
	}, nil
}

func (s *clinicService) UpdateClinic(ctx context.Context, req UpdateRequest, id, userID int) (ClinicResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ClinicResponse{}, err
	}

	if existing.UserID != userID {
		return ClinicResponse{}, ErrUnauthorizedAccess
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.PhoneNumber != "" {
		existing.PhoneNumber = req.PhoneNumber
	}
	if req.Address != "" {
		existing.Address = req.Address
	}
	if req.MapLink != "" {
		existing.MapLink = req.MapLink
	}

	if req.Image != nil {
		imageURL, err := s.uploadImage(ctx, req.Image)
		if err != nil {
			return ClinicResponse{}, err
		}
		existing.ImageURL = imageURL
	}

	if err := s.repo.Update(ctx, &existing); err != nil {
		s.logger.Error("Failed to update clinic", "id", id, "error", err)
		return ClinicResponse{}, ErrInvalidClinicData
	}

	return existing, nil
}

func (s *clinicService) DeleteClinic(ctx context.Context, id, userID int) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		s.logger.Error("Failed to delete clinic", "id", id, "error", err)
		return err
	}
	return nil
}

func (s *clinicService) SearchClinics(ctx context.Context, query string, page, pageSize int) (PaginationResponse, error) {
	if page < 1 || pageSize < 1 {
		return PaginationResponse{}, ErrPaginationInvalid
	}

	clinics, total, err := s.repo.Search(ctx, query, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to search clinics", "query", query, "error", err)
		return PaginationResponse{}, err
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	fromItem := (page-1)*pageSize + 1
	toItem := page * pageSize
	if toItem > int(total) {
		toItem = int(total)
	}

	return PaginationResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Message:    "Success",
		PageNumber: page,
		TotalPages: totalPages,
		FromItem:   fromItem,
		ToItem:     toItem,
		TotalItem:  total,
		Data:       clinics,
	}, nil
}

func (s *clinicService) uploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", ErrImageUploadFailed
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", ErrImageUploadFailed
	}

	mimeType := http.DetectContentType(fileBytes)
	if !isValidImageType(mimeType) {
		return "", ErrInvalidImageFormat
	}

	imageReader := bytes.NewReader(fileBytes)
	publicID := fmt.Sprintf("clinic_%d", time.Now().UnixNano())

	url, err := s.cloudinary.Upload(ctx, cloudinary.UploadParams{
		File:     imageReader,
		Folder:   "clinics",
		PublicID: publicID,
	})

	if err != nil {
		s.logger.Error("Image upload failed", "error", err)
		return "", ErrImageUploadFailed
	}

	return url, nil
}

func isValidImageType(mimeType string) bool {
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	return allowed[mimeType]
}
