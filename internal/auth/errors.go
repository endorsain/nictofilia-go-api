package auth

import (
	"net/http"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
)

var (
	ErrInvalidCredentials  = apperrors.New("INVALID_CREDENTIALS", "invalid credentials", http.StatusUnauthorized)
	ErrInvalidRefreshToken = apperrors.New("INVALID_REFRESH_TOKEN", "invalid refresh token", http.StatusUnauthorized)
)
