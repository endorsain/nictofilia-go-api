package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/auth"
	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
)

type contextKey struct{}

type Identity struct {
	UserID string
	Roles  []string
}

type Authentication struct {
	tokenService *auth.TokenService
}

func NewAuthentication(tokenService *auth.TokenService) *Authentication {
	return &Authentication{tokenService: tokenService}
}

func (m *Authentication) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			apperrors.Write(w, apperrors.New(apperrors.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		claims, err := m.tokenService.ValidateAccessToken(tokenString, time.Now())
		if err != nil || claims.Subject == "" {
			apperrors.Write(w, apperrors.New(apperrors.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		identity := Identity{
			UserID: claims.Subject,
			Roles:  append([]string(nil), claims.Roles...),
		}
		ctx := context.WithValue(r.Context(), contextKey{}, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(contextKey{}).(Identity)
	return identity, ok
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	identity, ok := IdentityFromContext(ctx)
	if !ok || identity.UserID == "" {
		return "", false
	}
	return identity.UserID, true
}

func RolesFromContext(ctx context.Context) ([]string, bool) {
	identity, ok := IdentityFromContext(ctx)
	if !ok {
		return nil, false
	}
	return append([]string(nil), identity.Roles...), true
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
