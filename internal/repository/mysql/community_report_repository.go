package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/traumacompendium/backend/internal/domain/community"
)

type CommunityReportRepository struct {
	db *sql.DB
}

func NewCommunityReportRepository(db *sql.DB) *CommunityReportRepository {
	return &CommunityReportRepository{db: db}
}

func (r *CommunityReportRepository) Create(ctx context.Context, rep *community.Report) error {
	const q = `
		INSERT INTO community_reports (id, target_type, target_id, reporter_user_id, reason, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, q,
		rep.ID, rep.TargetType, rep.TargetID, nullableString(rep.ReporterUserID), rep.Reason, rep.Status, rep.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("mysql: create community report: %w", err)
	}
	return nil
}

func (r *CommunityReportRepository) FindByID(ctx context.Context, id string) (*community.Report, error) {
	const q = `
		SELECT id, target_type, target_id, reporter_user_id, reason, status, resolved_by_user_id, resolved_at, created_at
		FROM community_reports
		WHERE id = ?
	`
	return scanCommunityReport(r.db.QueryRowContext(ctx, q, id))
}

func scanCommunityReport(row *sql.Row) (*community.Report, error) {
	var rep community.Report
	var reporterUserID, resolvedByUserID sql.NullString
	err := row.Scan(
		&rep.ID, &rep.TargetType, &rep.TargetID, &reporterUserID, &rep.Reason, &rep.Status,
		&resolvedByUserID, &rep.ResolvedAt, &rep.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, community.ErrReportNotFound
		}
		return nil, fmt.Errorf("mysql: find community report: %w", err)
	}
	rep.ReporterUserID = reporterUserID.String
	rep.ResolvedByUserID = resolvedByUserID.String
	return &rep, nil
}

func (r *CommunityReportRepository) ListOpen(ctx context.Context, filter community.ListFilter) ([]*community.Report, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultCommunityPageSize
	}

	const countQ = `SELECT COUNT(*) FROM community_reports WHERE status = ?`
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, community.ReportStatusOpen).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("mysql: count community reports: %w", err)
	}

	const listQ = `
		SELECT id, target_type, target_id, reporter_user_id, reason, status, resolved_by_user_id, resolved_at, created_at
		FROM community_reports
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, listQ, community.ReportStatusOpen, limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("mysql: list community reports: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only cursor, nothing actionable on close failure

	reports := make([]*community.Report, 0)
	for rows.Next() {
		var rep community.Report
		var reporterUserID, resolvedByUserID sql.NullString
		if err := rows.Scan(
			&rep.ID, &rep.TargetType, &rep.TargetID, &reporterUserID, &rep.Reason, &rep.Status,
			&resolvedByUserID, &rep.ResolvedAt, &rep.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("mysql: scan community report: %w", err)
		}
		rep.ReporterUserID = reporterUserID.String
		rep.ResolvedByUserID = resolvedByUserID.String
		reports = append(reports, &rep)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("mysql: iterate community reports: %w", err)
	}

	return reports, total, nil
}

func (r *CommunityReportRepository) Resolve(ctx context.Context, id, resolvedByUserID string) error {
	const q = `
		UPDATE community_reports
		SET status = ?, resolved_by_user_id = ?, resolved_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = ?
	`
	res, err := r.db.ExecContext(ctx, q, community.ReportStatusResolved, resolvedByUserID, id, community.ReportStatusOpen)
	if err != nil {
		return fmt.Errorf("mysql: resolve community report: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mysql: resolve community report rows affected: %w", err)
	}
	if rows == 0 {
		return community.ErrReportNotFound
	}
	return nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
