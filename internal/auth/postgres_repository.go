package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/endorsain/nictofilia-go-api/internal/users"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) FindLoginUser(ctx context.Context, username string) (users.User, error) {
	const query = `
		SELECT id, username, email, password_hash, created_at
		FROM users
		WHERE username = $1
	`

	var user users.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, users.ErrNotFound
		}
		return users.User{}, fmt.Errorf("find login user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) FindUserRoles(ctx context.Context, userID string) ([]string, error) {
	const query = `
		SELECT r.name
		FROM roles r
		INNER JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find user roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user roles: %w", err)
	}

	return roles, nil
}

func (r *PostgresRepository) SaveRefreshToken(ctx context.Context, token RefreshToken) error {
	const query = `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	if _, err := r.db.ExecContext(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}

	return nil
}

func (r *PostgresRepository) FindRefreshToken(ctx context.Context, tokenHash string) (RefreshTokenRecord, error) {
	const query = `
		SELECT id, user_id, expires_at, revoked
		FROM refresh_tokens
		WHERE token = $1
	`

	var token RefreshTokenRecord
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.ExpiresAt,
		&token.Revoked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshTokenRecord{}, ErrInvalidRefreshToken
		}
		return RefreshTokenRecord{}, fmt.Errorf("find refresh token: %w", err)
	}

	return token, nil
}

func (r *PostgresRepository) RotateRefreshToken(ctx context.Context, previousTokenID int64, token RefreshToken) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin refresh token rotation: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const revokeQuery = `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE id = $1
		  AND revoked = FALSE
		  AND expires_at > NOW()
	`

	result, err := tx.ExecContext(ctx, revokeQuery, previousTokenID)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check refresh token revocation: %w", err)
	}
	if rowsAffected != 1 {
		return ErrInvalidRefreshToken
	}

	const insertQuery = `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	if _, err = tx.ExecContext(ctx, insertQuery, token.UserID, token.TokenHash, token.ExpiresAt); err != nil {
		return fmt.Errorf("save rotated refresh token: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit refresh token rotation: %w", err)
	}

	return nil
}

func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE token = $1
		  AND revoked = FALSE
	`

	if _, err := r.db.ExecContext(ctx, query, tokenHash); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}
