package register

import "time"

// RegisterRequest digunakan untuk menerima data registrasi pengguna dari request JSON
type RegisterRequest struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Gender      string `json:"gender"`
	Username    string `json:"username" binding:"required"`
	DateOfBirth string `json:"date_of_birth" time_format:"2024-02-24"`
}

// User merepresentasikan data pengguna yang akan disimpan di database
type User struct {
	ID          int       `json:"id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	Gender      string    `json:"gender"`
	Username    string    `json:"username"`
	DateOfBirth string    `json:"date_of_birth"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
