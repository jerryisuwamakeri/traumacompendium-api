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

// Create and Update both persist u.CreatedAt/u.UpdatedAt as supplied by
// the caller rather than leaving them to the columns' DB-side defaults
// (DEFAULT CURRENT_TIMESTAMP / ON UPDATE CURRENT_TIMESTAMP): the service
// layer sets these from its own clock (Service.now) before calling in, so
// the struct the caller already holds stays accurate without needing a
// read-back, and the value is deterministic in tests.
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	const q = `
		INSERT INTO users (id, full_name, email, phone, password_hash, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q,
		u.ID, u.FullName, u.Email, u.Phone, u.PasswordHash, u.Role, u.Status, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return translateUserWriteError(err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	const q = `
		UPDATE users SET
			full_name = ?, email = ?, email_verified_at = ?, phone = ?, phone_verified_at = ?,
			password_hash = ?, role = ?, status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, q,
		u.FullName, u.Email, u.EmailVerifiedAt, u.Phone, u.PhoneVerifiedAt,
		u.PasswordHash, u.Role, u.Status, u.UpdatedAt, u.ID,
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

// translateUserWriteError turns a MySQL duplicate-key error on the users
// table into the matching domain sentinel (user.ErrEmailTaken /
// user.ErrPhoneTaken) so the service layer can react to it without
// importing this infrastructure package.
func translateUserWriteError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 { // ER_DUP_ENTRY
		switch {
		case strings.Contains(mysqlErr.Message, "uq_users_email"):
			return user.ErrEmailTaken
		case strings.Contains(mysqlErr.Message, "uq_users_phone"):
			return user.ErrPhoneTaken
		}
		return fmt.Errorf("%w", ErrDuplicateEntry)
	}
	return fmt.Errorf("mysql: write user: %w", err)
}
