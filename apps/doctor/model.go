package doctor

import (
	"time"
)

type DoctorRequest struct {
	UserID            int      `json:"user_id" binding:"required"`
	FullName          string   `json:"full_name" binding:"required"`
	LastEducation     string   `json:"last_education" binding:"required"`
	SpecialistAt      string   `json:"specialist_at" binding:"required"`
	ProfileImage      string   `json:"profile_image"`
	BirthDate         string   `json:"birth_date" binding:"required"`
	HospitalName      string   `json:"hospital_name"`
	YearsOfExperience int      `json:"years_of_experience"`
	PricePerHour      float64  `json:"price_per_hour" binding:"required"`
	GmeetLink         string   `json:"gmeet_link" binding:"required"`
	WorkingDays       []string `json:"working_days" binding:"required"`
	WorkingHoursStart string   `json:"working_hours_start" binding:"required"`
	WorkingHoursEnd   string   `json:"working_hours_end" binding:"required"`
}

type DoctorResponse struct {
	ID                int      `json:"id"`
	UserID            int      `json:"user_id" binding:"required"`
	FullName          string   `json:"full_name"`
	LastEducation     string   `json:"last_education"`
	SpecialistAt      string   `json:"specialist_at"`
	ProfileImage      string   `json:"profile_image"`
	BirthDate         string   `json:"birth_date" binding:"required"`
	HospitalName      string   `json:"hospital_name"`
	YearsOfExperience int      `json:"years_of_experience"`
	PricePerHour      float64  `json:"price_per_hour"`
	GmeetLink         string   `json:"gmeet_link"`
	WorkingDays       []string `json:"working_days"`
	WorkingHours      struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"working_hours"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Pagination struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}
