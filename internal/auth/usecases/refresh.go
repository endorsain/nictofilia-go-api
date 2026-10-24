package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/auth"
)

type RefreshTokenUseCase struct {
	authRepository auth.Repository
	tokenService   *auth.TokenService
}

type RefreshResult struct {
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

func NewRefreshTokenUseCase(authRepository auth.Repository, tokenService *auth.TokenService) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		authRepository: authRepository,
		tokenService:   tokenService,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, refreshToken string) (RefreshResult, error) {
	if refreshToken == "" {
		return RefreshResult{}, auth.ErrInvalidRefreshToken
	}

	refreshTokenRecord, err := uc.authRepository.FindRefreshToken(ctx, auth.HashRefreshToken(refreshToken))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return RefreshResult{}, auth.ErrInvalidRefreshToken
		}
		return RefreshResult{}, fmt.Errorf("find refresh token: %w", err)
	}

	if refreshTokenRecord.Revoked || !refreshTokenRecord.ExpiresAt.After(time.Now()) {
		return RefreshResult{}, auth.ErrInvalidRefreshToken
	}

	roles, err := uc.authRepository.FindUserRoles(ctx, refreshTokenRecord.UserID)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("find user roles: %w", err)
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(refreshTokenRecord.UserID, roles, time.Now())
	if err != nil {
		return RefreshResult{}, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, newRefreshTokenHash, newRefreshExpiresAt, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return RefreshResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := uc.authRepository.RotateRefreshToken(ctx, refreshTokenRecord.ID, auth.RefreshToken{
		UserID:    refreshTokenRecord.UserID,
		TokenHash: newRefreshTokenHash,
		ExpiresAt: newRefreshExpiresAt,
	}); err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			return RefreshResult{}, auth.ErrInvalidRefreshToken
		}
		return RefreshResult{}, fmt.Errorf("rotate refresh token: %w", err)
	}

	return RefreshResult{
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		RefreshExpiresAt: newRefreshExpiresAt,
	}, nil
}
