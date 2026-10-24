package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/endorsain/nictofilia-go-api/internal/users"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserUseCase struct {
	userRepository users.UserRepository
}

func NewRegisterUserUseCase(userRepository users.UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{userRepository: userRepository}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, username, email, password string) (users.User, error) {
	if username == "" || email == "" || password == "" {
		return users.User{}, users.ErrInvalidInput
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return users.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := users.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}
	profile := users.UserProfile{}

	createdUser, err := uc.userRepository.CreateDefaultUser(ctx, user, profile)
	if err != nil {
		if errors.Is(err, users.ErrUsernameTaken) {
			return users.User{}, users.ErrUsernameTaken
		}
		if errors.Is(err, users.ErrEmailTaken) {
			return users.User{}, users.ErrEmailTaken
		}
		return users.User{}, fmt.Errorf("create user: %w", err)
	}

	return createdUser, nil
}
