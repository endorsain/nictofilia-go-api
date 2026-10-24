package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdomain "github.com/endorsain/nictofilia-go-api/internal/auth"
	authmiddleware "github.com/endorsain/nictofilia-go-api/internal/auth/middleware"
	"github.com/endorsain/nictofilia-go-api/internal/config"
	"github.com/endorsain/nictofilia-go-api/internal/users"
	"github.com/go-chi/chi/v5"
)

type userWithProfileUseCaseStub struct {
	user users.UserWithProfile
}

func (s userWithProfileUseCaseStub) Execute(context.Context, string) (users.UserWithProfile, error) {
	return s.user, nil
}

type listUsersUseCaseStub struct{}

func (listUsersUseCaseStub) Execute(context.Context) ([]users.UserSummary, error) {
	return []users.UserSummary{{ID: "user-id", Username: "sample-user"}}, nil
}

func TestUserDetailEndpointsDoNotExposePasswordHash(t *testing.T) {
	tokenService, err := authdomain.NewTokenService(config.AuthConfig{
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
	accessToken, err := tokenService.GenerateAccessToken("user-id", nil, time.Now())
	if err != nil {
		t.Fatalf("generate test access token: %v", err)
	}

	user := users.UserWithProfile{
		User: users.User{
			ID:           "user-id",
			Username:     "sample-user",
			Email:        "sample-user@example.com",
			PasswordHash: "synthetic-test-hash",
			CreatedAt:    "2026-09-27T12:00:00Z",
		},
		Profile: users.UserProfile{
			ID:          "profile-id",
			UserID:      "user-id",
			DisplayName: "Sample User",
			Biography:   "Test biography",
			UpdatedAt:   "2026-09-27T12:00:00Z",
		},
	}

	tests := []struct {
		name string
		path string
		auth bool
	}{
		{name: "current user", path: "/users/me", auth: true},
		{name: "user by id", path: "/users/user-id"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := userWithProfileUseCaseStub{user: user}
			router := chi.NewRouter()
			requireAuthentication := authmiddleware.NewAuthentication(tokenService).RequireAuthentication
			RegisterRoutes(
				router,
				NewGetCurrentUserHandler(useCase),
				NewGetUserByIDHandler(useCase),
				NewListUsersHandler(listUsersUseCaseStub{}),
				requireAuthentication,
			)

			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.auth {
				request.Header.Set("Authorization", "Bearer "+accessToken)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("response status = %d, want %d", response.Code, http.StatusOK)
			}
			assertPublicUserJSON(t, response.Body.String(), user)
		})
	}
}

func assertPublicUserJSON(t *testing.T, body string, user users.UserWithProfile) {
	t.Helper()
	for _, forbidden := range []string{"PasswordHash", "password_hash", user.PasswordHash} {
		if strings.Contains(body, forbidden) {
			t.Fatal("users endpoint response contains a forbidden password field or value")
		}
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode users response JSON: %v", err)
	}
	if len(payload) != 5 {
		t.Fatalf("users response field count = %d, want 5 public fields", len(payload))
	}
	assertJSONField(t, payload, "ID", user.ID)
	assertJSONField(t, payload, "Username", user.Username)
	assertJSONField(t, payload, "Email", user.Email)
	assertJSONField(t, payload, "CreatedAt", user.CreatedAt)

	var profile map[string]json.RawMessage
	if err := json.Unmarshal(payload["Profile"], &profile); err != nil {
		t.Fatalf("decode public profile JSON: %v", err)
	}
	if len(profile) != 5 {
		t.Fatalf("profile response field count = %d, want 5 public fields", len(profile))
	}
	assertJSONField(t, profile, "ID", user.Profile.ID)
	assertJSONField(t, profile, "UserID", user.Profile.UserID)
	assertJSONField(t, profile, "DisplayName", user.Profile.DisplayName)
	assertJSONField(t, profile, "Biography", user.Profile.Biography)
	assertJSONField(t, profile, "UpdatedAt", user.Profile.UpdatedAt)
}

func assertJSONField(t *testing.T, object map[string]json.RawMessage, field, expected string) {
	t.Helper()
	value, ok := object[field]
	if !ok {
		t.Fatalf("response is missing public field %q", field)
	}
	var actual string
	if err := json.Unmarshal(value, &actual); err != nil {
		t.Fatalf("decode response field %q: %v", field, err)
	}
	if actual != expected {
		t.Fatalf("response field %q has unexpected value", field)
	}
}
