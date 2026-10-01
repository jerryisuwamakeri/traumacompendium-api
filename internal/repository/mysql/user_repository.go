// Package mysql implements every domain repository interface against
// MySQL using database/sql directly (no ORM) — hand-written SQL kept
// close to its table, with driver-specific concerns (duplicate-key
// detection, NULL handling) isolated here so the domain and service
// layers never import database/sql or the MySQL driver.
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"

	"github.com/traumacompendium/backend/internal/domain/user"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `
	id, full_name, email, email_verified_at, phone, phone_verified_at,
	password_hash, role, status, created_at, updated_at, deleted_at
`

func scanUser(row interface{ Scan(dest ...any) error }) (*user.User, error) {
	var u user.User
	err := row.Scan(
		&u.ID, &u.FullName, &u.Email, &u.EmailVerifiedAt, &u.Phone, &u.PhoneVerifiedAt,
		&u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (id, full_name, email, phone, password_hash, role, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, u.ID, u.FullName, u.Email, u.Phone, u.PasswordHash, u.Role, u.Status)
	if err != nil {
		return translateUserWriteError(err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	const q = `
		UPDATE users SET
			full_name = ?, email = ?, email_verified_at = ?, phone = ?, phone_verified_at = ?,
			password_hash = ?, role = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, q,
		u.FullName, u.Email, u.EmailVerifiedAt, u.Phone, u.PhoneVerifiedAt,
		u.PasswordHash, u.Role, u.Status, u.ID,
	)
	if err != nil {
		return translateUserWriteError(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: user update rows affected: %w", err)
	}
	if rows == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	const q = `SELECT ` + userColumns + ` FROM users WHERE id = ? AND deleted_at IS NULL`
	u, err := scanUser(r.db.QueryRowContext(ctx, q, id))
	return mapUserNotFound(u, err)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	const q = `SELECT ` + userColumns + ` FROM users WHERE email = ? AND deleted_at IS NULL`
	u, err := scanUser(r.db.QueryRowContext(ctx, q, strings.ToLower(email)))
	return mapUserNotFound(u, err)
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	const q = `SELECT ` + userColumns + ` FROM users WHERE phone = ? AND deleted_at IS NULL`
	u, err := scanUser(r.db.QueryRowContext(ctx, q, phone))
	return mapUserNotFound(u, err)
}

func mapUserNotFound(u *user.User, err error) (*user.User, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("mysql: query user: %w", err)
	}
	return u, nil
}

// translateUserWriteError turns a MySQL duplicate-key error on the
// users table into a domain-meaningful error the service layer can map to
// a 409 Conflict, instead of leaking a raw driver error.
func translateUserWriteError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 { // ER_DUP_ENTRY
		switch {
		case strings.Contains(mysqlErr.Message, "uq_users_email"):
			return fmt.Errorf("%w: email already registered", ErrDuplicateEntry)
		case strings.Contains(mysqlErr.Message, "uq_users_phone"):
			return fmt.Errorf("%w: phone already registered", ErrDuplicateEntry)
		}
		return fmt.Errorf("%w", ErrDuplicateEntry)
	}
	return fmt.Errorf("mysql: write user: %w", err)
}
