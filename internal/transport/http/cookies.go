package httptransport

import (
	"net/http"
	"strings"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/config"
)

func SetRefreshTokenCookie(w http.ResponseWriter, value string, expiresAt time.Time, cfg config.RefreshCookieConfig) {
	setRefreshTokenCookie(w, value, expiresAt, 0, cfg)
}

func ClearRefreshTokenCookie(w http.ResponseWriter, cfg config.RefreshCookieConfig) {
	setRefreshTokenCookie(w, "", time.Unix(0, 0), -1, cfg)
}

func setRefreshTokenCookie(w http.ResponseWriter, value string, expiresAt time.Time, maxAge int, cfg config.RefreshCookieConfig) {
	sameSite := http.SameSiteLaxMode
	switch strings.ToLower(cfg.SameSite) {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    value,
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		Secure:   cfg.Secure,
		HttpOnly: cfg.HTTPOnly,
		SameSite: sameSite,
		Expires:  expiresAt,
		MaxAge:   maxAge,
	})
}
