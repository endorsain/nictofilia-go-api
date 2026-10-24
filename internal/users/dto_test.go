package users

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestToUserWithProfileResponseExcludesPasswordHash(t *testing.T) {
	user := UserWithProfile{
		User: User{
			ID:           "user-id",
			Username:     "sample-user",
			Email:        "sample-user@example.com",
			PasswordHash: "synthetic-test-hash",
			CreatedAt:    "2026-09-27T12:00:00Z",
		},
		Profile: UserProfile{
			ID:          "profile-id",
			UserID:      "user-id",
			DisplayName: "Sample User",
			Biography:   "Test biography",
			UpdatedAt:   "2026-09-27T12:00:00Z",
		},
	}

	response := ToUserWithProfileResponse(user)
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response DTO: %v", err)
	}

	for _, forbidden := range []string{"PasswordHash", "password_hash", user.PasswordHash} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatal("response DTO contains a forbidden password field or value")
		}
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("decode response DTO: %v", err)
	}
	if len(payload) != 5 {
		t.Fatalf("response DTO field count = %d, want 5 public fields", len(payload))
	}
	for _, field := range []string{"ID", "Username", "Email", "CreatedAt", "Profile"} {
		if _, ok := payload[field]; !ok {
			t.Fatalf("response DTO is missing public field %q", field)
		}
	}
}

func TestUserModelOmitsPasswordHashFromJSON(t *testing.T) {
	user := User{ID: "user-id", PasswordHash: "synthetic-test-hash"}
	encoded, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal user model: %v", err)
	}
	if strings.Contains(string(encoded), "PasswordHash") || strings.Contains(string(encoded), "password_hash") || strings.Contains(string(encoded), user.PasswordHash) {
		t.Fatal("user model JSON contains a forbidden password field or value")
	}
}
