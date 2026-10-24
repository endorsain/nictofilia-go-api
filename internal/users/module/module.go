package module

import (
	"database/sql"
	"net/http"

	"github.com/endorsain/nictofilia-go-api/internal/users"
	usersHandlers "github.com/endorsain/nictofilia-go-api/internal/users/handlers"
	usersUsecases "github.com/endorsain/nictofilia-go-api/internal/users/usecases"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	userRepository        users.UserRepository
	getCurrentUserHandler *usersHandlers.GetCurrentUserHandler
	getUserByIDHandler    *usersHandlers.GetUserByIDHandler
	listUsersHandler      *usersHandlers.ListUsersHandler
}

func New(db *sql.DB) *Module {
	userRepository := users.NewPostgresUserRepository(db)
	getCurrentUserUseCase := usersUsecases.NewGetCurrentUserUseCase(userRepository)
	getUserByIDUseCase := usersUsecases.NewGetUserByIDUseCase(userRepository)
	listUsersUseCase := usersUsecases.NewListUsersUseCase(userRepository)

	return &Module{
		userRepository:        userRepository,
		getCurrentUserHandler: usersHandlers.NewGetCurrentUserHandler(getCurrentUserUseCase),
		getUserByIDHandler:    usersHandlers.NewGetUserByIDHandler(getUserByIDUseCase),
		listUsersHandler:      usersHandlers.NewListUsersHandler(listUsersUseCase),
	}
}

func (m *Module) RegisterRoutes(r chi.Router, requireAuthentication func(http.Handler) http.Handler) {
	usersHandlers.RegisterRoutes(r, m.getCurrentUserHandler, m.getUserByIDHandler, m.listUsersHandler, requireAuthentication)
}

func (m *Module) UserRepository() users.UserRepository {
	return m.userRepository
}
