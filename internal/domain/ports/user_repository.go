package ports

import (
	"context"
	"errors"

	"ahorrapp/internal/domain/entities"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("email already registered")

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, displayName string) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id string) (*entities.User, error)
	GetPoints(ctx context.Context, userID string) (int, error)
	RecentTransactions(ctx context.Context, userID string, limit int) ([]entities.LoyaltyTransaction, error)
}
