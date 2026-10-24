package handlers

import (
	"context"
	"fmt"
	"net/http"

	authmiddleware "github.com/endorsain/nictofilia-go-api/internal/auth/middleware"
	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	"github.com/endorsain/nictofilia-go-api/internal/users"
)

type GetCurrentUserUseCase interface {
	Execute(ctx context.Context, id string) (users.UserWithProfile, error)
}

type GetCurrentUserHandler struct {
	useCase GetCurrentUserUseCase
}

func NewGetCurrentUserHandler(useCase GetCurrentUserUseCase) *GetCurrentUserHandler {
	return &GetCurrentUserHandler{useCase: useCase}
}

func (h *GetCurrentUserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) error {
	userID, ok := authmiddleware.UserIDFromContext(r.Context())
	if !ok {
		return apperrors.New(apperrors.CodeUnauthorized, "unauthorized", http.StatusUnauthorized)
	}

	result, err := h.useCase.Execute(r.Context(), userID)
	if err != nil {
		return fmt.Errorf("get current user (id=%s): %w", userID, err)
	}

	httptransport.WriteJSON(w, http.StatusOK, users.ToUserWithProfileResponse(result))
	return nil
}
