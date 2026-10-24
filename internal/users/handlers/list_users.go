package handlers

import (
	"context"
	"fmt"
	"net/http"

	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	"github.com/endorsain/nictofilia-go-api/internal/users"
)

type ListUsersUseCase interface {
	Execute(ctx context.Context) ([]users.UserSummary, error)
}

type ListUsersHandler struct {
	useCase ListUsersUseCase
}

func NewListUsersHandler(useCase ListUsersUseCase) *ListUsersHandler {
	return &ListUsersHandler{useCase: useCase}
}

func (h *ListUsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) error {
	result, err := h.useCase.Execute(r.Context())
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	httptransport.WriteJSON(w, http.StatusOK, result)
	return nil
}
