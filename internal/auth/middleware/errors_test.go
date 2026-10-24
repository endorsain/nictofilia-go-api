package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/auth"
	"github.com/endorsain/nictofilia-go-api/internal/config"
	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
)

type publicErrorEnvelope struct {
	Error struct {
		Code    apperrors.Code `json:"code"`
		Message string         `json:"message"`
	} `json:"error"`
}

func TestAuthenticationReturnsCentralEnvelopeForFailures(t *testing.T) {
	tokenService := newTestTokenService(t)
	expiredToken, err := tokenService.GenerateAccessToken("user-id", nil, time.Now().Add(-2*time.Minute))
	if err != nil {
		t.Fatalf("generate expired test token: %v", err)
	}

	invalidToken := "synthetic-invalid-token"
	tests := []struct {
		name       string
		authorize  string
		wantStatus int
		leakCheck  string
	}{
		{name: "missing authorization", wantStatus: http.StatusUnauthorized},
		{name: "malformed bearer header", authorize: "Basic credentials", wantStatus: http.StatusUnauthorized},
		{name: "empty bearer token", authorize: "Bearer ", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", authorize: "Bearer " + invalidToken, wantStatus: http.StatusUnauthorized, leakCheck: invalidToken},
		{name: "expired token", authorize: "Bearer " + expiredToken, wantStatus: http.StatusUnauthorized, leakCheck: expiredToken},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if test.authorize != "" {
				request.Header.Set("Authorization", test.authorize)
			}
			response := httptest.NewRecorder()
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})

			NewAuthentication(tokenService).RequireAuthentication(next).ServeHTTP(response, request)

			assertPublicError(t, response, test.wantStatus, apperrors.CodeUnauthorized, "unauthorized", test.leakCheck)
			for _, privateDetail := range []string{"signature", "expired", "unit-test-signing-secret"} {
				if strings.Contains(response.Body.String(), privateDetail) {
					t.Fatalf("authentication response leaked %q", privateDetail)
				}
			}
		})
	}
}

func TestAuthenticationContinuesWithValidToken(t *testing.T) {
	tokenService := newTestTokenService(t)
	accessToken, err := tokenService.GenerateAccessToken("user-id", []string{"regular"}, time.Now())
	if err != nil {
		t.Fatalf("generate test access token: %v", err)
	}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		called = true
		identity, ok := IdentityFromContext(request.Context())
		if !ok || identity.UserID != "user-id" || len(identity.Roles) != 1 || identity.Roles[0] != "regular" {
			t.Errorf("identity in context = %#v, present = %v", identity, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()

	NewAuthentication(tokenService).RequireAuthentication(next).ServeHTTP(response, request)

	if !called {
		t.Fatal("next handler was not called for a valid token")
	}
	if response.Code != http.StatusNoContent {
		t.Fatalf("response status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestRequireRoleReturnsCentralErrorsAndAllowsMatchingRole(t *testing.T) {
	tests := []struct {
		name       string
		identity   *Identity
		wantStatus int
		wantCode   apperrors.Code
		wantMsg    string
	}{
		{
			name:       "missing identity",
			wantStatus: http.StatusUnauthorized,
			wantCode:   apperrors.CodeUnauthorized,
			wantMsg:    "unauthorized",
		},
		{
			name:       "missing role",
			identity:   &Identity{UserID: "user-id", Roles: []string{"regular"}},
			wantStatus: http.StatusForbidden,
			wantCode:   apperrors.CodeForbidden,
			wantMsg:    "forbidden",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if test.identity != nil {
				request = request.WithContext(context.WithValue(request.Context(), contextKey{}, *test.identity))
			}
			response := httptest.NewRecorder()
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})

			RequireRole("admin")(next).ServeHTTP(response, request)

			assertPublicError(t, response, test.wantStatus, test.wantCode, test.wantMsg)
		})
	}

	t.Run("matching role continues", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		request = request.WithContext(context.WithValue(request.Context(), contextKey{}, Identity{
			UserID: "user-id",
			Roles:  []string{"admin"},
		}))
		response := httptest.NewRecorder()
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		})

		RequireRole("admin")(next).ServeHTTP(response, request)

		if !called || response.Code != http.StatusNoContent {
			t.Fatalf("authorized request did not continue: called=%v status=%d", called, response.Code)
		}
	})
}

func TestMiddlewareAndHandlerUseSamePublicErrorEnvelope(t *testing.T) {
	appError := apperrors.New(apperrors.CodeUnauthorized, "unauthorized", http.StatusUnauthorized)
	middlewareResponse := httptest.NewRecorder()
	apperrors.Write(middlewareResponse, appError)

	handlerResponse := httptest.NewRecorder()
	httptransport.Handle(func(http.ResponseWriter, *http.Request) error { return appError })(
		handlerResponse,
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	if middlewareResponse.Code != handlerResponse.Code {
		t.Fatalf("middleware status = %d, handler status = %d", middlewareResponse.Code, handlerResponse.Code)
	}
	if middlewareResponse.Header().Get("Content-Type") != handlerResponse.Header().Get("Content-Type") {
		t.Fatalf("middleware content type = %q, handler content type = %q", middlewareResponse.Header().Get("Content-Type"), handlerResponse.Header().Get("Content-Type"))
	}
	if middlewareResponse.Body.String() != handlerResponse.Body.String() {
		t.Fatalf("middleware envelope = %q, handler envelope = %q", middlewareResponse.Body.String(), handlerResponse.Body.String())
	}
}

func newTestTokenService(t *testing.T) *auth.TokenService {
	t.Helper()
	tokenService, err := auth.NewTokenService(config.AuthConfig{
		JWT: config.JWTConfig{
			Secret:   "unit-test-signing-secret",
			Issuer:   "unit-test-issuer",
			Audience: "unit-test-audience",
			TTL:      time.Minute,
		},
	})
	if err != nil {
		t.Fatalf("create token service: %v", err)
	}
	return tokenService
}

func assertPublicError(t *testing.T, response *httptest.ResponseRecorder, status int, code apperrors.Code, message string, forbiddenValues ...string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("response status = %d, want %d", response.Code, status)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var envelope publicErrorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode public error envelope: %v", err)
	}
	if envelope.Error.Code != code || envelope.Error.Message != message {
		t.Fatalf("public error = %#v, want code=%q message=%q", envelope.Error, code, message)
	}
	if strings.Contains(response.Body.String(), `"error":"`) {
		t.Fatalf("response uses legacy flat error format: %s", response.Body.String())
	}
	for _, value := range forbiddenValues {
		if value != "" && strings.Contains(response.Body.String(), value) {
			t.Fatalf("response leaked sensitive value %q", value)
		}
	}
}
