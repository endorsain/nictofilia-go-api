package auth

import (
	"context"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/users"
)

type Repository interface {
	FindLoginUser(ctx context.Context, username string) (users.User, error)
	FindUserRoles(ctx context.Context, userID string) ([]string, error)
	SaveRefreshToken(ctx context.Context, token RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (RefreshTokenRecord, error)
	RotateRefreshToken(ctx context.Context, previousTokenID int64, token RefreshToken) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}

type RefreshToken struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
}

type RefreshTokenRecord struct {
	ID        int64
	UserID    string
	ExpiresAt time.Time
	Revoked   bool
}
