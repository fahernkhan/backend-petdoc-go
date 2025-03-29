package consultation

import (
	"mime/multipart"
	"time"
)

// Request dari FE
type CreateRequest struct {
	UserID             int                   `form:"user_id" binding:"required"`
	DoctorID           int                   `form:"doctor_id" binding:"required"`
	PetType            string                `form:"pet_type" binding:"required"`
	PetName            string                `form:"pet_name" binding:"required"`
	PetAge             int                   `form:"pet_age" binding:"required,min=0,max=30"`
	DiseaseDescription string                `form:"disease_description" binding:"required,min=10,max=2000"`
	ConsultationDate   string                `form:"consultation_date" binding:"required"` // Format: YYYY-MM-DD
	StartTime          string                `form:"start_time" binding:"required"`        // Format: 2006-01-02T15:04:05Z
	EndTime            string                `form:"end_time" binding:"required"`          // Format: 2006-01-02T15:04:05Z
	PaymentProof       *multipart.FileHeader `form:"payment_proof" binding:"required"`
}

// Response untuk FE
type ConsultationResponse struct {
	ID                 int       `json:"id"`
	UserID             int       `json:"user_id"`
	DoctorID           int       `json:"doctor_id"`
	PetType            string    `json:"pet_type"`
	PetName            string    `json:"pet_name"`
	PetAge             int       `json:"pet_age"`
	DiseaseDescription string    `json:"disease_description"`
	ConsultationDate   string    `json:"consultation_date"` // Format: "YYYY-MM-DD"
	StartTimeUTC       string    `json:"start_time_utc"`    // Format RFC3339
	EndTimeUTC         string    `json:"end_time_utc"`      // Format RFC3339
	StartTimeWIB       string    `json:"start_time_wib"`    // Format: "2006-01-02 15:04:05"
	EndTimeWIB         string    `json:"end_time_wib"`      // Format: "2006-01-02 15:04:05"
	PaymentProof       string    `json:"payment_proof"`
	CreatedAt          time.Time `json:"created_at"`
	MeetLink           string    `json:"meet_link"`
}

// Untuk Pagination
type PaginationResponse struct {
	Data       []ConsultationResponse `json:"data"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalItems int64                  `json:"total_items"`
	TotalPages int                    `json:"total_pages"`
}

type DoctorSchedule struct {
	GMeetLink    string
	Price        float64
	WorkingDays  []string
	WorkingHours struct {
		Start string
		End   string
	}
}
