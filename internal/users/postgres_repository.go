package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (User, error) {
	const query = `
		SELECT id, username, email, password_hash
		FROM users
		WHERE username = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("find user by username: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id, username, email, password_hash
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user User, profile UserProfile) (User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	const insertUserQuery = `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err = tx.QueryRowContext(ctx, insertUserQuery, user.Username, user.Email, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		return User{}, err
	}

	profile.UserID = user.ID
	const insertProfileQuery = `
		INSERT INTO user_profiles (user_id)
		VALUES ($1)
	`

	_, err = tx.ExecContext(ctx, insertProfileQuery, profile.UserID)
	if err != nil {
		return User{}, err
	}

	if err = tx.Commit(); err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) CreateDefaultUser(ctx context.Context, user User, profile UserProfile) (_ User, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const insertUserQuery = `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err = tx.QueryRowContext(ctx, insertUserQuery, user.Username, user.Email, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		return User{}, mapCreateUserError(err)
	}

	profile.UserID = user.ID
	const insertProfileQuery = `
		INSERT INTO user_profiles (user_id)
		VALUES ($1)
	`

	_, err = tx.ExecContext(ctx, insertProfileQuery, profile.UserID)
	if err != nil {
		return User{}, fmt.Errorf("insert user profile: %w", err)
	}

	const defaultRoleIDQuery = `
		SELECT id
		FROM roles
		WHERE name = 'regular'
	`

	var roleID int
	err = tx.QueryRowContext(ctx, defaultRoleIDQuery).Scan(&roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, fmt.Errorf("default role not found: %w", ErrNotFound)
		}
		return User{}, fmt.Errorf("find default role: %w", err)
	}

	const assignRoleQuery = `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
	`

	_, err = tx.ExecContext(ctx, assignRoleQuery, user.ID, roleID)
	if err != nil {
		return User{}, fmt.Errorf("assign default role: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}

func mapCreateUserError(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return fmt.Errorf("insert user: %w", err)
	}

	if pqErr.Code == "23505" {
		switch pqErr.Constraint {
		case "users_username_key":
			return ErrUsernameTaken
		case "users_email_key":
			return ErrEmailTaken
		}

		detail := strings.ToLower(pqErr.Detail)
		if strings.Contains(detail, "(username)=") {
			return ErrUsernameTaken
		}
		if strings.Contains(detail, "(email)=") {
			return ErrEmailTaken
		}
	}

	return fmt.Errorf("insert user: %w", err)
}

// TODO: mejorar el logging de errores.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (UserWithProfile, error) {
	const query = `
		SELECT
		    u.id, u.username, u.email, u.created_at,
			p.id, p.user_id, COALESCE(p.display_name, ''), COALESCE(p.biography, ''), p.updated_at
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE u.id = $1
	`

	var user UserWithProfile
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.Profile.ID,
		&user.Profile.UserID,
		&user.Profile.DisplayName,
		&user.Profile.Biography,
		&user.Profile.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserWithProfile{}, ErrNotFound
		}
		return UserWithProfile{}, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) List(ctx context.Context) ([]UserSummary, error) {
	const query = `
		SELECT id, username
		FROM users
		ORDER BY username ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []UserSummary
	for rows.Next() {
		var user UserSummary
		err = rows.Scan(&user.ID, &user.Username)
		if err != nil {
			return nil, fmt.Errorf("scan user summary: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}
