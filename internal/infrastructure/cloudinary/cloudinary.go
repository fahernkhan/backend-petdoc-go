package cloudinary

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type Service interface {
	Upload(ctx context.Context, params UploadParams) (string, error)
}

type service struct {
	cli *cloudinary.Cloudinary
}

type UploadParams struct {
	File     string // Base64 string
	Folder   string
	PublicID string
}

func NewService(cloudName, apiKey, apiSecret string) (Service, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi Cloudinary: %w", err)
	}
	return &service{cli: cld}, nil
}

func (s *service) Upload(ctx context.Context, params UploadParams) (string, error) {
	// Decode base64
	data := params.File
	if i := strings.Index(data, ","); i != -1 {
		data = data[i+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("gagal decode base64: %w", err)
	}

	// Upload ke Cloudinary
	overwrite := true
	result, err := s.cli.Upload.Upload(ctx, decoded, uploader.UploadParams{
		Folder:    params.Folder,
		PublicID:  fmt.Sprintf("%s_%d", params.PublicID, time.Now().Unix()),
		Overwrite: &overwrite,
	})
	if err != nil {
		return "", fmt.Errorf("gagal upload ke Cloudinary: %w", err)
	}

	return result.SecureURL, nil
}
