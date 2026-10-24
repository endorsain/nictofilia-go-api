package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	authModule "github.com/endorsain/nictofilia-go-api/internal/auth/module"
	"github.com/endorsain/nictofilia-go-api/internal/config"
	"github.com/endorsain/nictofilia-go-api/internal/infrastructure/postgres"
	"github.com/endorsain/nictofilia-go-api/internal/logging"
	httptransport "github.com/endorsain/nictofilia-go-api/internal/transport/http"
	usersModule "github.com/endorsain/nictofilia-go-api/internal/users/module"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.Logging)

	logger.Info("configuration loaded",
		"app", cfg.Application.Name,
		"env", cfg.Application.Environment,
		"server", cfg.Server.Address(),
		"db_host", cfg.Database.Host,
		"db_port", cfg.Database.Port,
		"db_name", cfg.Database.Name,
		"db_user", cfg.Database.User,
		"db_sslmode", cfg.Database.SSLMode,
	)

	pgClient, err := postgres.New(context.Background(), cfg.Database, logger)
	if err != nil {
		logger.Error("failed to connect to PostgreSQL", "error", err.Error())
		os.Exit(1)
	}

	users := usersModule.New(pgClient.DB())
	requestValidator := httptransport.NewRequestValidator()
	auth, err := authModule.New(pgClient.DB(), users.UserRepository(), cfg.Auth, requestValidator)
	if err != nil {
		logger.Error("failed to initialize auth module", "error", err.Error())
		_ = pgClient.Close()
		os.Exit(1)
	}

	httptransport.SetErrorLogger(logger)

	router := httptransport.NewRouter(cfg.Server)
	auth.RegisterRoutes(router)
	users.RegisterRoutes(router, auth.RequireAuthentication)

	server := httptransport.NewServer(cfg.Server, router)

	logger.Info("starting server")
	logger.Info("HTTP server starting", "address", server.Address())

	serverErrCh := make(chan error, 1)
	go func() {
		err := server.Start()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		close(serverErrCh)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrCh:
		if err != nil {
			logger.Error("HTTP server stopped with error", "error", err.Error())
		}
	case sig := <-signalCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server stopped with error", "error", err.Error())
	}

	if err := pgClient.Close(); err != nil {
		logger.Error("failed to close PostgreSQL connection", "error", err.Error())
	}
}
