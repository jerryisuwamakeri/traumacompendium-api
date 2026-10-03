package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/certificate"
)

type CertificateRepository struct {
	db *sql.DB
}

func NewCertificateRepository(db *sql.DB) *CertificateRepository {
	return &CertificateRepository{db: db}
}

func (r *CertificateRepository) Create(ctx context.Context, c *certificate.Certificate) error {
	const q = `
		INSERT INTO certificates (id, user_id, level_id, serial, issued_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q, c.ID, c.UserID, c.LevelID, c.Serial, c.IssuedAt)
	if err != nil {
		return fmt.Errorf("mysql: create certificate: %w", err)
	}
	return nil
}

func (r *CertificateRepository) FindByUserAndLevel(ctx context.Context, userID, levelID string) (certificate.Certificate, error) {
	const q = `
		SELECT id, user_id, level_id, serial, issued_at
		FROM certificates
		WHERE user_id = ? AND level_id = ?
	`
	return scanCertificate(r.db.QueryRowContext(ctx, q, userID, levelID))
}

func (r *CertificateRepository) FindBySerial(ctx context.Context, serial string) (certificate.Certificate, error) {
	const q = `
		SELECT id, user_id, level_id, serial, issued_at
		FROM certificates
		WHERE serial = ?
	`
	return scanCertificate(r.db.QueryRowContext(ctx, q, serial))
}

func scanCertificate(row *sql.Row) (certificate.Certificate, error) {
	var c certificate.Certificate
	err := row.Scan(&c.ID, &c.UserID, &c.LevelID, &c.Serial, &c.IssuedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return certificate.Certificate{}, certificate.ErrNotFound
		}
		return certificate.Certificate{}, fmt.Errorf("mysql: find certificate: %w", err)
	}
	return c, nil
}
