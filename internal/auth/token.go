package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/endorsain/nictofilia-go-api/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	jwtConfig          config.JWTConfig
	refreshTokenConfig config.RefreshTokenConfig
}

type AccessTokenClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewTokenService(authConfig config.AuthConfig) (*TokenService, error) {
	if authConfig.JWT.Secret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	return &TokenService{
		jwtConfig:          authConfig.JWT,
		refreshTokenConfig: authConfig.RefreshToken,
	}, nil
}

func (s *TokenService) GenerateAccessToken(userID string, roles []string, now time.Time) (string, error) {
	issuedAt := now.UTC()
	claims := AccessTokenClaims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.jwtConfig.Issuer,
			Audience:  jwt.ClaimStrings{s.jwtConfig.Audience},
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(s.jwtConfig.TTL)),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

func (s *TokenService) ValidateAccessToken(tokenString string, now time.Time) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return []byte(s.jwtConfig.Secret), nil
	}, jwt.WithIssuer(s.jwtConfig.Issuer), jwt.WithAudience(s.jwtConfig.Audience), jwt.WithTimeFunc(func() time.Time {
		return now
	}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return claims, nil
}

func (s *TokenService) GenerateRefreshToken() (value string, hash string, expiresAt time.Time, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}

	value = base64.RawURLEncoding.EncodeToString(bytes)
	digest := sha256.Sum256([]byte(value))
	hash = hex.EncodeToString(digest[:])
	expiresAt = time.Now().UTC().Add(s.refreshTokenConfig.TTL)
	return value, hash, expiresAt, nil
}

func HashRefreshToken(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
