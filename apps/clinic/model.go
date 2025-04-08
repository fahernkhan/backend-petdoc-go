package clinic

import (
	"mime/multipart"
	"time"
)

type CreateRequest struct {
	Name        string                `form:"name" binding:"required,min=3,max=255"`
	PhoneNumber string                `form:"phone_number" binding:"required,min=8,max=20"`
	Address     string                `form:"address" binding:"required,min=10"`
	MapLink     string                `form:"map_link"`
	Image       *multipart.FileHeader `form:"image" binding:"required"`
}

type UpdateRequest struct {
	Name        string                `form:"name" binding:"omitempty,min=3,max=255"`
	PhoneNumber string                `form:"phone_number" binding:"omitempty,min=8,max=20"`
	Address     string                `form:"address" binding:"omitempty,min=10"`
	MapLink     string                `form:"map_link"`
	Image       *multipart.FileHeader `form:"image"`
}

type ClinicResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phone_number"`
	Address     string    `json:"address"`
	MapLink     string    `json:"map_link"`
	ImageURL    string    `json:"image_url"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaginationResponse struct {
	Success    bool             `json:"success"`
	StatusCode int              `json:"statusCode"`
	Message    string           `json:"message"`
	PageNumber int              `json:"pageNumber"`
	TotalPages int              `json:"totalPages"`
	FromItem   int              `json:"fromItem"`
	ToItem     int              `json:"toItem"`
	TotalItem  int64            `json:"totalItem"`
	Data       []ClinicResponse `json:"data"`
}
