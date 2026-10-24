package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/endorsain/nictofilia-go-api/internal/config"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
)

type LogoutUseCase interface {
	Execute(ctx context.Context, refreshToken string) error
}

type LogoutHandler struct {
	useCase      LogoutUseCase
	cookieConfig config.RefreshCookieConfig
}

func NewLogoutHandler(useCase LogoutUseCase, cookieConfig config.RefreshCookieConfig) *LogoutHandler {
	return &LogoutHandler{useCase: useCase, cookieConfig: cookieConfig}
}

// Logout is idempotent: a missing, already revoked or unknown refresh token
// still results in a successful logout. Only an unexpected failure is
// reported to the centralized error middleware.
func (h *LogoutHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(h.cookieConfig.Name)
	if err == nil {
		if err := h.useCase.Execute(r.Context(), cookie.Value); err != nil {
			return fmt.Errorf("logout: %w", err)
		}
	} else if err != http.ErrNoCookie {
		return fmt.Errorf("read refresh token cookie: %w", err)
	}

	httptransport.ClearRefreshTokenCookie(w, h.cookieConfig)
	w.WriteHeader(http.StatusNoContent)
	return nil
}
