package journalservice

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func seedTwoEntries(t *testing.T, svc *Service) {
	t.Helper()
	ctx := context.Background()
	if _, err := svc.Create(ctx, "user-1", CreateInput{Title: "First", Content: "one"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Create(ctx, "user-1", CreateInput{Title: "Second", Content: "two"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestExport_JSON(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	seedTwoEntries(t, svc)

	data, contentType, err := svc.Export(context.Background(), "user-1", ExportFormatJSON)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if contentType != "application/json" {
		t.Errorf("contentType = %q, want application/json", contentType)
	}

	var entries []ExportEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
}

func TestExport_CSV(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	seedTwoEntries(t, svc)

	data, contentType, err := svc.Export(context.Background(), "user-1", ExportFormatCSV)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if contentType != "text/csv" {
		t.Errorf("contentType = %q, want text/csv", contentType)
	}

	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatalf("decode csv: %v", err)
	}
	// header + 2 entries
	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3 (header + 2 entries)", len(rows))
	}
	if rows[0][0] != "id" {
		t.Errorf("header row = %v, want it to start with 'id'", rows[0])
	}
}

func TestExport_UnsupportedFormat_ReturnsBadRequest(t *testing.T) {
	svc := New(fakes.NewJournalRepo())

	_, _, err := svc.Export(context.Background(), "user-1", ExportFormat("xml"))
	if err == nil {
		t.Fatal("Export() with unsupported format error = nil, want error")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error = %v, want it to mention the unsupported format", err)
	}
}

func TestExport_OnlyIncludesCallersEntries(t *testing.T) {
	svc := New(fakes.NewJournalRepo())
	ctx := context.Background()

	seedTwoEntries(t, svc)
	if _, err := svc.Create(ctx, "user-2", CreateInput{Title: "Someone else's", Content: "..."}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	data, _, err := svc.Export(ctx, "user-1", ExportFormatJSON)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	var entries []ExportEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2 (not including user-2's entry)", len(entries))
	}
}
