package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/endorsain/nictofilia-go-api/internal/auth/usecases"
	"github.com/endorsain/nictofilia-go-api/internal/config"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
)

type LoginUseCase interface {
	Execute(ctx context.Context, username, password string) (usecases.LoginResult, error)
}

type LoginHandler struct {
	useCase          LoginUseCase
	cookieConfig     config.RefreshCookieConfig
	requestValidator *httptransport.RequestValidator
}

type loginRequest struct {
	Username string `json:"username" validate:"required,max=50"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

func NewLoginHandler(useCase LoginUseCase, cookieConfig config.RefreshCookieConfig, requestValidator *httptransport.RequestValidator) *LoginHandler {
	return &LoginHandler{useCase: useCase, cookieConfig: cookieConfig, requestValidator: requestValidator}
}

func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req loginRequest
	if err := httptransport.DecodeJSON(r, &req); err != nil {
		return err
	}
	if err := h.requestValidator.Validate(req); err != nil {
		return err
	}

	result, err := h.useCase.Execute(r.Context(), req.Username, req.Password)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	httptransport.SetRefreshTokenCookie(w, result.RefreshToken, result.RefreshExpiresAt, h.cookieConfig)
	httptransport.WriteJSON(w, http.StatusOK, loginResponse{AccessToken: result.AccessToken})
	return nil
}
