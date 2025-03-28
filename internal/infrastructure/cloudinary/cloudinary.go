package cloudinary

import (
	"context"
	"fmt"
	"io"
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
	File     io.Reader // Ubah ke io.Reader
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
	overwrite := true

	// Log parameter upload
	fmt.Printf("[DEBUG] Upload Params - Folder: %s, PublicID: %s\n",
		params.Folder,
		params.PublicID,
	)

	result, err := s.cli.Upload.Upload(ctx, params.File, uploader.UploadParams{
		Folder:    params.Folder,
		PublicID:  fmt.Sprintf("%s_%d", params.PublicID, time.Now().Unix()),
		Overwrite: &overwrite,
	})

	if err != nil {
		// Log error lengkap
		fmt.Printf("[ERROR] Cloudinary Upload Failed: %v\n", err)
		return "", fmt.Errorf("gagal upload ke Cloudinary: %w", err)
	}

	// Log hasil sukses
	fmt.Printf("[DEBUG] Upload Success - URL: %s\n", result.SecureURL)
	return result.SecureURL, nil
}

// package cloudinary

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/cloudinary/cloudinary-go/v2"
// 	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
// )

// type Service interface {
// 	Upload(ctx context.Context, params UploadParams) (string, error)
// }

// type service struct {
// 	cli *cloudinary.Cloudinary
// }

// type UploadParams struct {
// 	File     string // Ubah kembali ke string untuk base64
// 	Folder   string
// 	PublicID string
// }

// func NewService(cloudName, apiKey, apiSecret string) (Service, error) {
// 	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
// 	if err != nil {
// 		return nil, fmt.Errorf("gagal inisialisasi Cloudinary: %w", err)
// 	}
// 	return &service{cli: cld}, nil
// }

// func (s *service) Upload(ctx context.Context, params UploadParams) (string, error) {
// 	overwrite := true
// 	// Handle base64 string dengan benar
// 	result, err := s.cli.Upload.Upload(ctx, params.File, uploader.UploadParams{
// 		Folder:    params.Folder,
// 		PublicID:  fmt.Sprintf("%s_%d", params.PublicID, time.Now().Unix()),
// 		Overwrite: &overwrite,
// 	})

// 	if err != nil {
// 		return "", fmt.Errorf("gagal upload ke Cloudinary: %w", err)
// 	}

// 	return result.SecureURL, nil
// }

// // func (s *service) Upload(ctx context.Context, params UploadParams) (string, error) {
// // 	// Decode base64
// // 	data := params.File
// // 	if i := strings.Index(data, ","); i != -1 {
// // 		data = data[i+1:]
// // 	}
// // 	// decoded, err := base64.StdEncoding.DecodeString(data)
// // 	// if err != nil {
// // 	// 	return "", fmt.Errorf("gagal decode base64: %w", err)
// // 	// }

// // 	// Upload ke Cloudinary
// // 	overwrite := true
// // 	// Langsung gunakan byte file tanpa decode base64
// // 	result, err := s.cli.Upload.Upload(ctx, params.File, uploader.UploadParams{ // params.File harus []byte
// // 		Folder:    params.Folder,
// // 		PublicID:  fmt.Sprintf("%s_%d", params.PublicID, time.Now().Unix()),
// // 		Overwrite: &overwrite,
// // 	})
// // 	if err != nil {
// // 		return "", fmt.Errorf("gagal upload ke Cloudinary: %w", err)
// // 	}

// // 	return result.SecureURL, nil
// // }
