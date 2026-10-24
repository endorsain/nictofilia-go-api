package module

import (
	"database/sql"
	"net/http"

	"github.com/endorsain/nictofilia-go-api/internal/auth"
	"github.com/endorsain/nictofilia-go-api/internal/auth/handlers"
	"github.com/endorsain/nictofilia-go-api/internal/auth/middleware"
	"github.com/endorsain/nictofilia-go-api/internal/auth/usecases"
	"github.com/endorsain/nictofilia-go-api/internal/config"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	"github.com/endorsain/nictofilia-go-api/internal/users"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	registerHandler *handlers.RegisterHandler
	loginHandler    *handlers.LoginHandler
	refreshHandler  *handlers.RefreshHandler
	logoutHandler   *handlers.LogoutHandler
	authentication  *middleware.Authentication
}

func New(db *sql.DB, userRepository users.UserRepository, authConfig config.AuthConfig, requestValidator *httptransport.RequestValidator) (*Module, error) {
	registerUseCase := usecases.NewRegisterUserUseCase(userRepository)
	authRepository := auth.NewPostgresRepository(db)
	tokenService, err := auth.NewTokenService(authConfig)
	if err != nil {
		return nil, err
	}
	loginUseCase := usecases.NewLoginUseCase(authRepository, tokenService)
	refreshTokenUseCase := usecases.NewRefreshTokenUseCase(authRepository, tokenService)
	logoutUseCase := usecases.NewLogoutUseCase(authRepository)

	return &Module{
		registerHandler: handlers.NewHandler(registerUseCase, requestValidator),
		loginHandler:    handlers.NewLoginHandler(loginUseCase, authConfig.RefreshCookie, requestValidator),
		refreshHandler:  handlers.NewRefreshHandler(refreshTokenUseCase, authConfig.RefreshCookie),
		logoutHandler:   handlers.NewLogoutHandler(logoutUseCase, authConfig.RefreshCookie),
		authentication:  middleware.NewAuthentication(tokenService),
	}, nil
}

func (m *Module) RegisterRoutes(r chi.Router) {
	handlers.RegisterRoutes(r, m.registerHandler, m.loginHandler, m.refreshHandler, m.logoutHandler)
}

func (m *Module) RequireAuthentication(next http.Handler) http.Handler {
	return m.authentication.RequireAuthentication(next)
}
