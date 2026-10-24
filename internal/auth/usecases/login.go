package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/auth"
	"github.com/endorsain/nictofilia-go-api/internal/users"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	authRepository auth.Repository
	tokenService   *auth.TokenService
}

type LoginResult struct {
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

func NewLoginUseCase(authRepository auth.Repository, tokenService *auth.TokenService) *LoginUseCase {
	return &LoginUseCase{
		authRepository: authRepository,
		tokenService:   tokenService,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, username, password string) (LoginResult, error) {
	user, err := uc.authRepository.FindLoginUser(ctx, username)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return LoginResult{}, auth.ErrInvalidCredentials
		}
		return LoginResult{}, fmt.Errorf("find login user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, auth.ErrInvalidCredentials
	}

	roles, err := uc.authRepository.FindUserRoles(ctx, user.ID)
	if err != nil {
		return LoginResult{}, fmt.Errorf("find user roles: %w", err)
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(user.ID, roles, time.Now())
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, refreshTokenHash, refreshExpiresAt, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := uc.authRepository.SaveRefreshToken(ctx, auth.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return LoginResult{}, fmt.Errorf("save refresh token: %w", err)
	}

	return LoginResult{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}
