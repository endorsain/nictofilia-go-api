package middleware

import (
	"bytes"
	stdErrors "errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
)

func TestErrorsLogsOnlyUnexpectedServerErrors(t *testing.T) {
	var logOutput bytes.Buffer
	previousLogger := logger
	SetLogger(slog.New(slog.NewJSONHandler(&logOutput, nil)))
	defer SetLogger(previousLogger)

	clientError := apperrors.New(apperrors.CodeValidationError, "request validation failed", http.StatusBadRequest)
	clientResponse := httptest.NewRecorder()
	Errors(func(http.ResponseWriter, *http.Request) error { return clientError })(clientResponse, httptest.NewRequest(http.MethodPost, "/", nil))
	if logOutput.Len() != 0 {
		t.Fatalf("4xx error was logged: %s", logOutput.String())
	}
	if clientResponse.Code != http.StatusBadRequest {
		t.Fatalf("4xx response status = %d, want %d", clientResponse.Code, http.StatusBadRequest)
	}

	serverResponse := httptest.NewRecorder()
	Errors(func(http.ResponseWriter, *http.Request) error { return stdErrors.New("database connection details") })(serverResponse, httptest.NewRequest(http.MethodGet, "/", nil))
	if serverResponse.Code != http.StatusInternalServerError {
		t.Fatalf("5xx response status = %d, want %d", serverResponse.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(logOutput.String(), "request failed") {
		t.Fatalf("5xx error was not logged: %s", logOutput.String())
	}
	if strings.Contains(serverResponse.Body.String(), "database connection details") {
		t.Fatalf("internal error leaked in response: %s", serverResponse.Body.String())
	}
}
