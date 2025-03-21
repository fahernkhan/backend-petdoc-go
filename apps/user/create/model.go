package user

import "time"

type User struct {
	ID           int
	FullName     string
	Email        string
	Password     string
	Gender       string
	username     string
	DateOfBirth  time.Time
	Role         string
	ProviderID   string
	ProviderName string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
