package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/auth/usecases"
	"github.com/endorsain/nictofilia-go-api/internal/config"
	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	"github.com/endorsain/nictofilia-go-api/internal/users"
)

type registerUseCaseStub struct {
	called bool
}

func (s *registerUseCaseStub) Execute(context.Context, string, string, string) (users.User, error) {
	s.called = true
	return users.User{ID: "user-id", Username: "sample-user", Email: "sample-user@example.com"}, nil
}

type loginUseCaseStub struct {
	called bool
}

func (s *loginUseCaseStub) Execute(context.Context, string, string) (usecases.LoginResult, error) {
	s.called = true
	return usecases.LoginResult{
		AccessToken:      "access-token",
		RefreshToken:     "opaque-refresh-token",
		RefreshExpiresAt: time.Now().Add(time.Hour),
	}, nil
}

func TestRegisterHandlerRejectsInvalidRequestsBeforeUseCase(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode apperrors.Code
	}{
		{name: "malformed JSON", body: `{`, wantCode: apperrors.CodeInvalidInput},
		{name: "invalid DTO", body: `{"username":"sample-user","email":"bad","password":"secret"}`, wantCode: apperrors.CodeValidationError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := &registerUseCaseStub{}
			handler := NewHandler(useCase, httptransport.NewRequestValidator())
			request := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(test.body))
			response := httptest.NewRecorder()

			err := handler.RegisterUser(response, request)
			appErr, ok := apperrors.As(err)
			if !ok || appErr.Code != test.wantCode {
				t.Fatalf("handler error = %v, want AppError code %q", err, test.wantCode)
			}
			if useCase.called {
				t.Fatal("use case was called for invalid request")
			}
		})
	}
}

func TestRegisterHandlerPassesValidRequestToUseCase(t *testing.T) {
	useCase := &registerUseCaseStub{}
	handler := NewHandler(useCase, httptransport.NewRequestValidator())
	request := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"username":"sample-user","email":"sample-user@example.com","password":"password123"}`))
	response := httptest.NewRecorder()

	if err := handler.RegisterUser(response, request); err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}
	if !useCase.called {
		t.Fatal("use case was not called for valid request")
	}
	if response.Code != http.StatusCreated {
		t.Fatalf("response status = %d, want %d", response.Code, http.StatusCreated)
	}
}

func TestLoginHandlerValidatesRequestBeforeUseCase(t *testing.T) {
	useCase := &loginUseCaseStub{}
	handler := NewLoginHandler(useCase, config.RefreshCookieConfig{Name: "refresh_token"}, httptransport.NewRequestValidator())
	request := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"","password":"secret"}`))
	response := httptest.NewRecorder()

	err := handler.Login(response, request)
	appErr, ok := apperrors.As(err)
	if !ok || appErr.Code != apperrors.CodeValidationError {
		t.Fatalf("handler error = %v, want VALIDATION_ERROR", err)
	}
	if useCase.called {
		t.Fatal("use case was called for invalid request")
	}
}

func TestLoginHandlerPassesValidRequestToUseCase(t *testing.T) {
	useCase := &loginUseCaseStub{}
	handler := NewLoginHandler(useCase, config.RefreshCookieConfig{Name: "refresh_token"}, httptransport.NewRequestValidator())
	request := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"sample-user","password":"password123"}`))
	response := httptest.NewRecorder()

	if err := handler.Login(response, request); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !useCase.called {
		t.Fatal("use case was not called for valid request")
	}
	if response.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d", response.Code, http.StatusOK)
	}
}
