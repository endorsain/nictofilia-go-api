package httptransport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
	"github.com/endorsain/nictofilia-go-api/internal/users"
)

func TestRequestValidatorValidatesRegistrationDTO(t *testing.T) {
	requestValidator := NewRequestValidator()
	validRequest := users.RegisterUserRequest{
		Username: "sample-user",
		Email:    "sample-user@example.com",
		Password: "password123",
	}

	if err := requestValidator.Validate(validRequest); err != nil {
		t.Fatalf("Validate(valid request) error = %v", err)
	}

	tests := []struct {
		name    string
		request users.RegisterUserRequest
	}{
		{
			name: "required fields",
		},
		{
			name: "invalid email",
			request: users.RegisterUserRequest{
				Username: "sample-user",
				Email:    "not-an-email",
				Password: "password123",
			},
		},
		{
			name: "username exceeds database limit",
			request: users.RegisterUserRequest{
				Username: strings.Repeat("u", 51),
				Email:    "sample-user@example.com",
				Password: "password123",
			},
		},
		{
			name: "email exceeds database limit",
			request: users.RegisterUserRequest{
				Username: "sample-user",
				Email:    strings.Repeat("a", 244) + "@example.com",
				Password: "password123",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := requestValidator.Validate(test.request)
			if err == nil {
				t.Fatal("Validate() error = nil, want validation error")
			}

			appErr, ok := apperrors.As(err)
			if !ok {
				t.Fatalf("Validate() error type = %T, want *AppError", err)
			}
			if appErr.Code != apperrors.CodeValidationError {
				t.Fatalf("AppError.Code = %q, want %q", appErr.Code, apperrors.CodeValidationError)
			}
			if appErr.HTTPStatus != http.StatusBadRequest {
				t.Fatalf("AppError.HTTPStatus = %d, want %d", appErr.HTTPStatus, http.StatusBadRequest)
			}
			if err.Error() != "request validation failed" {
				t.Fatalf("public validation error = %q, want generic message", err.Error())
			}

			status, response := apperrors.Response(err)
			if status != http.StatusBadRequest || response.Error.Code != apperrors.CodeValidationError || response.Error.Message != "request validation failed" {
				t.Fatalf("public response = (%d, %#v), want generic validation envelope", status, response)
			}
		})
	}
}

func TestDecodeJSONReturnsControlledErrorForMalformedBodies(t *testing.T) {
	for _, body := range []string{"{", `{"username":12}`} {
		t.Run(body, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			var destination users.RegisterUserRequest
			err := DecodeJSON(request, &destination)
			if err == nil {
				t.Fatal("DecodeJSON() error = nil, want invalid JSON error")
			}

			status, response := apperrors.Response(err)
			if status != http.StatusBadRequest || response.Error.Code != apperrors.CodeInvalidInput || response.Error.Message != "invalid JSON body" {
				t.Fatalf("public response = (%d, %#v), want controlled invalid-input envelope", status, response)
			}
		})
	}
}

func TestDecodeJSONDecodesValidBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"sample-user","email":"sample-user@example.com","password":"password123"}`))
	var destination users.RegisterUserRequest
	if err := DecodeJSON(request, &destination); err != nil {
		t.Fatalf("DecodeJSON() error = %v", err)
	}
	if destination.Username != "sample-user" || destination.Email != "sample-user@example.com" {
		t.Fatalf("decoded request = %#v", destination)
	}
}
