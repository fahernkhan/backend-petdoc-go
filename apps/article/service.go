package article

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
	CreateArticle(ctx context.Context, req CreateRequest, userID int) (Response, error)
	GetArticle(ctx context.Context, id int) (Response, error)
	GetAllArticles(ctx context.Context, query string, page, pageSize int) (PaginationResponse, error)
	UpdateArticle(ctx context.Context, req UpdateRequest, id, userID int) (Response, error)
	DeleteArticle(ctx context.Context, id, userID int) error
	SearchArticles(ctx context.Context, query string, page, pageSize int) (PaginationResponse, error)
}

type articleService struct {
	repo       Repository
	cloudinary cloudinary.Service
	logger     *slog.Logger
}

func NewService(repo Repository, cloudinary cloudinary.Service, logger *slog.Logger) Service {
	return &articleService{
		repo:       repo,
		cloudinary: cloudinary,
		logger:     logger,
	}
}

func (s *articleService) CreateArticle(ctx context.Context, req CreateRequest, userID int) (Response, error) {
	imageURL, err := s.uploadImage(ctx, req.Image)
	if err != nil {
		return Response{}, err
	}

	article := Response{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		ImageURL:    imageURL,
		UserID:      userID,
		PublishedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, &article); err != nil {
		s.logger.Error("Failed to create article", "error", err)
		return Response{}, ErrInvalidArticleData
	}

	return article, nil
}

func (s *articleService) GetArticle(ctx context.Context, id int) (Response, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get article", "id", id, "error", err)
		return Response{}, ErrArticleNotFound
	}
	return article, nil
}

func (s *articleService) GetAllArticles(ctx context.Context, query string, page, pageSize int) (PaginationResponse, error) {
	if page < 1 || pageSize < 1 {
		return PaginationResponse{}, ErrPaginationInvalid
	}

	var articles []Response
	var total int64
	var err error

	if query != "" {
		articles, total, err = s.repo.Search(ctx, query, page, pageSize)
	} else {
		articles, total, err = s.repo.GetAll(ctx, page, pageSize)
	}

	if err != nil {
		s.logger.Error("Failed to get articles", "error", err)
		return PaginationResponse{}, err
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return PaginationResponse{
		Data:       articles,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (s *articleService) UpdateArticle(ctx context.Context, req UpdateRequest, id, userID int) (Response, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Response{}, err
	}

	if existing.UserID != userID {
		return Response{}, ErrUnauthorizedAccess
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Category != "" {
		existing.Category = req.Category
	}

	if req.Image != nil {
		imageURL, err := s.uploadImage(ctx, req.Image)
		if err != nil {
			return Response{}, err
		}
		existing.ImageURL = imageURL
	}

	if err := s.repo.Update(ctx, &existing); err != nil {
		s.logger.Error("Failed to update article", "id", id, "error", err)
		return Response{}, ErrInvalidArticleData
	}

	return existing, nil
}

func (s *articleService) DeleteArticle(ctx context.Context, id, userID int) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		s.logger.Error("Failed to delete article", "id", id, "error", err)
		return err
	}
	return nil
}

func (s *articleService) SearchArticles(ctx context.Context, query string, page, pageSize int) (PaginationResponse, error) {
	if page < 1 || pageSize < 1 {
		return PaginationResponse{}, ErrPaginationInvalid
	}

	articles, total, err := s.repo.Search(ctx, query, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to search articles", "query", query, "error", err)
		return PaginationResponse{}, err
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return PaginationResponse{
		Data:       articles,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (s *articleService) uploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", ErrImageUploadFailed
	}
	defer func() {
		closeErr := src.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", ErrImageUploadFailed
	}

	// Validate MIME type
	mimeType := http.DetectContentType(fileBytes)
	if !isValidImageType(mimeType) {
		return "", ErrInvalidImageFormat
	}

	imageReader := bytes.NewReader(fileBytes)
	publicID := fmt.Sprintf("article_%d", time.Now().UnixNano())

	url, err := s.cloudinary.Upload(ctx, cloudinary.UploadParams{
		File:     imageReader,
		Folder:   "articles",
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
