package journalservice

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
)

// ExportEntry is the shape written into an export — deliberately a
// separate type from journal.Entry so the export file format can evolve
// independently of the internal domain model.
type ExportEntry struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Export returns a user's complete journal history (brief: "exportable
// progress history") as either JSON or CSV, along with the MIME type to
// serve it with.
func (s *Service) Export(ctx context.Context, userID string, format ExportFormat) ([]byte, string, error) {
	entries, err := s.entries.ListAll(ctx, userID)
	if err != nil {
		return nil, "", apperror.NewInternal("could not export journal").WithCause(err)
	}

	exportEntries := make([]ExportEntry, len(entries))
	for i, e := range entries {
		exportEntries[i] = ExportEntry{
			ID:        e.ID,
			Title:     e.Title,
			Content:   e.Content,
			CreatedAt: e.CreatedAt,
			UpdatedAt: e.UpdatedAt,
		}
	}

	switch format {
	case ExportFormatCSV:
		return exportCSV(exportEntries)
	case ExportFormatJSON, "":
		return exportJSON(exportEntries)
	default:
		return nil, "", apperror.NewBadRequest(fmt.Sprintf("unsupported export format %q", format))
	}
}

func exportJSON(entries []ExportEntry) ([]byte, string, error) {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return nil, "", apperror.NewInternal("could not encode export").WithCause(err)
	}
	return data, "application/json", nil
}

func exportCSV(entries []ExportEntry) ([]byte, string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write([]string{"id", "title", "content", "created_at", "updated_at"}); err != nil {
		return nil, "", apperror.NewInternal("could not encode export").WithCause(err)
	}

	for _, e := range entries {
		row := []string{
			e.ID, e.Title, e.Content,
			e.CreatedAt.Format(time.RFC3339),
			e.UpdatedAt.Format(time.RFC3339),
		}
		if err := w.Write(row); err != nil {
			return nil, "", apperror.NewInternal("could not encode export").WithCause(err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", apperror.NewInternal("could not encode export").WithCause(err)
	}

	return buf.Bytes(), "text/csv", nil
}
