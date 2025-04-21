package admin

import (
	"time"
)

type CreateAdminRequest struct {
	UserID int `json:"user_id" binding:"required,min=1"`
}

type AdminResponse struct {
	Message      string    `json:"message"`
	UserID       int       `json:"user_id"`
	PreviousRole string    `json:"previous_role"`
	NewRole      string    `json:"new_role"`
	PromotedAt   time.Time `json:"promoted_at"`
}
