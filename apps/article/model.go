package article

import (
	"mime/multipart"
	"time"
)

type CreateRequest struct {
	Title       string                `form:"title" binding:"required,min=3,max=255"`
	Description string                `form:"description" binding:"required,min=10"`
	Category    string                `form:"category" binding:"required,min=3"`
	Image       *multipart.FileHeader `form:"image" binding:"required"`
}

type UpdateRequest struct {
	Title       string                `form:"title" binding:"omitempty,min=3,max=255"`
	Description string                `form:"description" binding:"omitempty,min=10"`
	Category    string                `form:"category" binding:"omitempty,min=3"`
	Image       *multipart.FileHeader `form:"image"`
}

type Response struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	UserID      int       `json:"user_id"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaginationResponse struct {
	Data       []Response `json:"data"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalItems int64      `json:"total_items"`
	TotalPages int        `json:"total_pages"`
}
