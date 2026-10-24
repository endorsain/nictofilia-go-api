package handlers

import (
	"context"
	"fmt"
	"net/http"

	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	"github.com/endorsain/nictofilia-go-api/internal/users"
)

type RegisterUserUseCase interface {
	Execute(ctx context.Context, username, email, password string) (users.User, error)
}

type RegisterHandler struct {
	usecase          RegisterUserUseCase
	requestValidator *httptransport.RequestValidator
}

func NewHandler(registerUserUseCase RegisterUserUseCase, requestValidator *httptransport.RequestValidator) *RegisterHandler {
	return &RegisterHandler{usecase: registerUserUseCase, requestValidator: requestValidator}
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) error {
	var req users.RegisterUserRequest
	if err := httptransport.DecodeJSON(r, &req); err != nil {
		return err
	}
	if err := h.requestValidator.Validate(req); err != nil {
		return err
	}

	createdUser, err := h.usecase.Execute(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		return fmt.Errorf("register user: %w", err)
	}

	httptransport.WriteJSON(w, http.StatusCreated, users.UserResponse{
		ID:       createdUser.ID,
		Username: createdUser.Username,
		Email:    createdUser.Email,
	})
	return nil
}
