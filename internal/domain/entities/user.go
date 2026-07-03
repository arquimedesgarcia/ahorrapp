package entities

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	Points       int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type LoyaltyTransaction struct {
	ID        string
	UserID    string
	Points    int
	Reason    string
	CreatedAt time.Time
	ReceiptID *string
}
