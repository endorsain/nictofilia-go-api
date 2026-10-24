package usecases

import (
	"context"

	"github.com/endorsain/nictofilia-go-api/internal/users"
)

type GetUserByIDUseCase struct {
	userRepository users.UserRepository
}

func NewGetUserByIDUseCase(userRepository users.UserRepository) *GetUserByIDUseCase {
	return &GetUserByIDUseCase{userRepository: userRepository}
}

func (uc *GetUserByIDUseCase) Execute(ctx context.Context, id string) (users.UserWithProfile, error) {
	return uc.userRepository.FindByID(ctx, id)
}
