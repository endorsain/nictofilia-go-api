package handlers

import (
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, registerHandler *RegisterHandler, loginHandler *LoginHandler, refreshHandler *RefreshHandler, logoutHandler *LogoutHandler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", httptransport.Handle(registerHandler.RegisterUser))
		r.Post("/login", httptransport.Handle(loginHandler.Login))
		r.Post("/refresh", httptransport.Handle(refreshHandler.Refresh))
		r.Post("/logout", httptransport.Handle(logoutHandler.Logout))
	})
}
