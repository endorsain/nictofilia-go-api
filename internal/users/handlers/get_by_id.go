package handlers

import (
	"context"
	"fmt"
	"net/http"

	apperrors "github.com/endorsain/nictofilia-go-api/internal/errors"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	"github.com/endorsain/nictofilia-go-api/internal/users"
	"github.com/go-chi/chi/v5"
)

type GetUserByIDUseCase interface {
	Execute(ctx context.Context, id string) (users.UserWithProfile, error)
}

type GetUserByIDHandler struct {
	useCase GetUserByIDUseCase
}

func NewGetUserByIDHandler(useCase GetUserByIDUseCase) *GetUserByIDHandler {
	return &GetUserByIDHandler{useCase: useCase}
}

func (h *GetUserByIDHandler) GetUserByID(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if id == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "id is required", http.StatusBadRequest)
	}

	result, err := h.useCase.Execute(r.Context(), id)
	if err != nil {
		return fmt.Errorf("get user by id (id=%s): %w", id, err)
	}

	httptransport.WriteJSON(w, http.StatusOK, users.ToUserWithProfileResponse(result))
	return nil
}
