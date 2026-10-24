package errors

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the public representation of an error, safe to expose to API
// clients. It never contains internal details (SQL, JWT, stack traces, ...).
type ErrorBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

// Envelope is the top-level JSON shape returned to clients:
//
//	{ "error": { "code": "...", "message": "..." } }
type Envelope struct {
	Error ErrorBody `json:"error"`
}

// HTTPStatus resolves the HTTP status code associated with err. Unknown
// errors default to 500, since they are treated as internal errors.
func HTTPStatus(err error) int {
	if appErr, ok := As(err); ok && appErr.HTTPStatus != 0 {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// Response builds the public HTTP status and body for err. Any error that is
// not an *AppError is treated as an unexpected internal error and reduced to
// a generic, safe envelope.
func Response(err error) (int, Envelope) {
	appErr, ok := As(err)
	if !ok {
		appErr = Internal(err)
	}

	return appErr.HTTPStatus, Envelope{
		Error: ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	}
}

// Write resolves err to its public HTTP representation and writes it as JSON.
// It is meant to be used by HTTP middleware/handlers, never exposing the
// original wrapped error to the client.
func Write(w http.ResponseWriter, err error) {
	status, body := Response(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
