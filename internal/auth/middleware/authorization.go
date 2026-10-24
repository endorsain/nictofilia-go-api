package middleware

import (
	"net/http"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
)

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, ok := RolesFromContext(r.Context())
			if !ok {
				apperrors.Write(w, apperrors.New(apperrors.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
				return
			}

			for _, currentRole := range roles {
				if currentRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			apperrors.Write(w, apperrors.New(apperrors.CodeForbidden, "forbidden", http.StatusForbidden))
		})
	}
}
