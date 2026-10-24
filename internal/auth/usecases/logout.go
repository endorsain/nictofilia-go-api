package usecases

import (
	"context"
	"fmt"

	"github.com/endorsain/nictofilia-go-api/internal/auth"
)

type LogoutUseCase struct {
	authRepository auth.Repository
}

func NewLogoutUseCase(authRepository auth.Repository) *LogoutUseCase {
	return &LogoutUseCase{authRepository: authRepository}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	if err := uc.authRepository.RevokeRefreshToken(ctx, auth.HashRefreshToken(refreshToken)); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}
