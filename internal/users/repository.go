package users

import "context"

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	CreateDefaultUser(ctx context.Context, user User, profile UserProfile) (User, error)
	Create(ctx context.Context, user User, profile UserProfile) (User, error)
	FindByID(ctx context.Context, id string) (UserWithProfile, error)
	List(ctx context.Context) ([]UserSummary, error)
}
