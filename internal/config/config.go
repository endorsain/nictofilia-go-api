package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Application ApplicationConfig
	Database    DatabaseConfig
	Server      ServerConfig
	Auth        AuthConfig
	Logging     LoggingConfig
}

type ApplicationConfig struct {
	Name        string
	Environment string
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	CORS            CORSConfig
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type AuthConfig struct {
	JWT           JWTConfig
	RefreshToken  RefreshTokenConfig
	RefreshCookie RefreshCookieConfig
}

type JWTConfig struct {
	Secret   string
	Issuer   string
	Audience string
	TTL      time.Duration
}

type RefreshTokenConfig struct {
	TTL time.Duration
}

type RefreshCookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	HTTPOnly bool
	SameSite string
}

type LoggingConfig struct {
	Level string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Application: ApplicationConfig{
			Name:        getenvOrDefault("APP_NAME", "nictofilia-go-api"),
			Environment: getenvOrDefault("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			URL:      os.Getenv("DATABASE_URL"),
			Host:     getenvOrDefault("DB_HOST", "localhost"),
			Port:     getenvOrDefault("DB_PORT", "5432"),
			Name:     getenvOrDefault("DB_NAME", "nictofilia"),
			User:     getenvOrDefault("DB_USER", "postgres"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  getenvOrDefault("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Host:            getenvOrDefault("SERVER_HOST", ""),
			Port:            getenvOrDefault("SERVER_PORT", "8080"),
			ReadTimeout:     durationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    durationOrDefault("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     durationOrDefault("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: durationOrDefault("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			CORS: CORSConfig{
				AllowedOrigins:   splitAndTrim(getenvOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
				AllowedMethods:   splitAndTrim(getenvOrDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")),
				AllowedHeaders:   splitAndTrim(getenvOrDefault("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")),
				AllowCredentials: boolOrDefault("CORS_ALLOW_CREDENTIALS", true),
			},
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				Secret:   os.Getenv("JWT_SECRET"),
				Issuer:   getenvOrDefault("JWT_ISSUER", "nictofilia-go-api"),
				Audience: getenvOrDefault("JWT_AUDIENCE", "nictofilia-go-api"),
				TTL:      durationOrDefault("JWT_TTL", 15*time.Minute),
			},
			RefreshToken: RefreshTokenConfig{
				TTL: durationOrDefault("REFRESH_TOKEN_TTL", 30*24*time.Hour),
			},
			RefreshCookie: RefreshCookieConfig{
				Name:     getenvOrDefault("REFRESH_COOKIE_NAME", "refresh_token"),
				Path:     getenvOrDefault("REFRESH_COOKIE_PATH", "/auth"),
				Domain:   os.Getenv("REFRESH_COOKIE_DOMAIN"),
				Secure:   boolOrDefault("REFRESH_COOKIE_SECURE", false),
				HTTPOnly: boolOrDefault("REFRESH_COOKIE_HTTP_ONLY", true),
				SameSite: strings.ToLower(getenvOrDefault("REFRESH_COOKIE_SAME_SITE", "lax")),
			},
		},
		Logging: LoggingConfig{
			Level: getenvOrDefault("LOG_LEVEL", "info"),
		},
	}
}

func (c DatabaseConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

func (c ServerConfig) Address() string {
	if c.Host == "" {
		return ":" + c.Port
	}
	return c.Host + ":" + c.Port
}

func getenvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// splitAndTrim splits a comma-separated env value into trimmed, non-empty parts.
func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func durationOrDefault(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func boolOrDefault(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
