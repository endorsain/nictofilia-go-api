package users

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string `json:"-"`
	CreatedAt    string
}

type UserProfile struct {
	ID          string
	UserID      string
	DisplayName string
	Biography   string
	UpdatedAt   string
}

type UserWithProfile struct {
	User
	Profile UserProfile
}

type UserSummary struct {
	ID       string
	Username string
}
