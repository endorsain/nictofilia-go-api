// Package middleware provides HTTP middleware for the transport layer that
// is generic across domains, such as centralized error handling.
package middleware

import (
	"log/slog"
	"net/http"
	"os"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
)

// HandlerFunc is an HTTP handler that reports failures by returning an error
// instead of writing the response itself.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// logger is used to record unexpected/internal errors. It defaults to a
// JSON logger so the middleware is safe to use even before SetLogger is
// called from the application composition root.
var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// SetLogger configures the logger used to record request failures.
func SetLogger(l *slog.Logger) {
	if l != nil {
		logger = l
	}
}

// Errors adapts a HandlerFunc into a standard http.HandlerFunc, centralizing
// error handling for the whole API:
//   - the original error (with its wrapping context) is logged once, here,
//     for the server; it is never sent to the client;
//   - the error is mapped to a stable HTTP status/code/message via
//     internal/errors, using errors.Is/errors.As under the hood so domain
//     sentinel errors (users.ErrNotFound, auth.ErrInvalidCredentials, ...)
//     are recognized regardless of how they were wrapped.
//   - anything that is not a recognized *errors.AppError is treated as an
//     unexpected internal error and reduced to a generic 500 response.
func Errors(next HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)
		if err == nil {
			return
		}

		if apperrors.HTTPStatus(err) >= http.StatusInternalServerError {
			logger.Error("request failed",
				"error", err.Error(),
				"request_id", chimiddleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
			)
		}
		apperrors.Write(w, err)
	}
}
