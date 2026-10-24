package handlers

import (
	"net/http"

	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, getCurrentUserHandler *GetCurrentUserHandler, getUserByIDHandler *GetUserByIDHandler, listUsersHandler *ListUsersHandler, requireAuthentication func(http.Handler) http.Handler) {
	r.Route("/users", func(r chi.Router) {
		r.With(requireAuthentication).Get("/me", httptransport.Handle(getCurrentUserHandler.GetCurrentUser))
		r.Get("/", httptransport.Handle(listUsersHandler.ListUsers))
		r.Get("/{id}", httptransport.Handle(getUserByIDHandler.GetUserByID))
	})
}
