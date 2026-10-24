package users

import (
	"net/http"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
)

var (
	ErrNotFound      = apperrors.New("USER_NOT_FOUND", "user not found", http.StatusNotFound)
	ErrInvalidInput  = apperrors.New("INVALID_USER_INPUT", "invalid user input", http.StatusBadRequest)
	ErrUsernameTaken = apperrors.New("USERNAME_ALREADY_EXISTS", "username already exists", http.StatusConflict)
	ErrEmailTaken    = apperrors.New("EMAIL_ALREADY_EXISTS", "email already exists", http.StatusConflict)
)
