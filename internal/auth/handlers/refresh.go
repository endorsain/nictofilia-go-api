package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	authDomain "github.com/endorsain/nictofilia-go-api/internal/auth"
	"github.com/endorsain/nictofilia-go-api/internal/auth/usecases"
	"github.com/endorsain/nictofilia-go-api/internal/config"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
)

type RefreshTokenUseCase interface {
	Execute(ctx context.Context, refreshToken string) (usecases.RefreshResult, error)
}

type RefreshHandler struct {
	useCase      RefreshTokenUseCase
	cookieConfig config.RefreshCookieConfig
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

func NewRefreshHandler(useCase RefreshTokenUseCase, cookieConfig config.RefreshCookieConfig) *RefreshHandler {
	return &RefreshHandler{useCase: useCase, cookieConfig: cookieConfig}
}

func (h *RefreshHandler) Refresh(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(h.cookieConfig.Name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return authDomain.ErrInvalidRefreshToken
		}
		return fmt.Errorf("read refresh token cookie: %w", err)
	}

	result, err := h.useCase.Execute(r.Context(), cookie.Value)
	if err != nil {
		return fmt.Errorf("refresh token: %w", err)
	}

	httptransport.SetRefreshTokenCookie(w, result.RefreshToken, result.RefreshExpiresAt, h.cookieConfig)
	httptransport.WriteJSON(w, http.StatusOK, refreshResponse{AccessToken: result.AccessToken})
	return nil
}
