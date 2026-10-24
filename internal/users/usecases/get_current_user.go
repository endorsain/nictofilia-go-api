package usecases

import (
	"context"

	"github.com/endorsain/nictofilia-go-api/internal/users"
)

type GetCurrentUserUseCase struct {
	userRepository users.UserRepository
}

func NewGetCurrentUserUseCase(userRepository users.UserRepository) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{userRepository: userRepository}
}

func (uc *GetCurrentUserUseCase) Execute(ctx context.Context, id string) (users.UserWithProfile, error) {
	return uc.userRepository.FindByID(ctx, id)
}
