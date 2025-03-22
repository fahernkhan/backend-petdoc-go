package login

import (
	"time"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserResponse struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        UserResponse `json:"user"`
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

// OR u Can == User menggunakan model dari register untuk konsistensi
// type User = register.User
