package pdf

import (
	"bytes"
	"testing"
	"time"
)

func TestNewCertificate_ProducesWellFormedPDF(t *testing.T) {
	issuedAt := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	out, err := NewCertificate("Ada Lovelace", "Self-Awareness", issuedAt, "TC-ABCD1234")
	if err != nil {
		t.Fatalf("NewCertificate() error = %v", err)
	}

	if !bytes.HasPrefix(out, []byte("%PDF-1.4")) {
		t.Errorf("output does not start with the PDF header, got first bytes: %q", out[:min(20, len(out))])
	}
	if !bytes.HasSuffix(bytes.TrimRight(out, "\n"), []byte("%%EOF")) {
		t.Errorf("output does not end with %%%%EOF, got last bytes: %q", out[max(0, len(out)-20):])
	}

	for _, want := range []string{
		"Certificate of Completion",
		"Ada Lovelace",
		"Self-Awareness",
		"Issued March 15, 2026",
		"Certificate No. TC-ABCD1234",
		"/BaseFont /Helvetica",
		"/BaseFont /Helvetica-Bold",
		"xref",
		"trailer",
	} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("output missing expected content %q", want)
		}
	}
}

func TestNewCertificate_EscapesSpecialCharacters(t *testing.T) {
	out, err := NewCertificate("O'Brien (Test)", "Level \\Six\\", time.Now(), "TC-XYZ")
	if err != nil {
		t.Fatalf("NewCertificate() error = %v", err)
	}
	if !bytes.Contains(out, []byte(`O'Brien \(Test\)`)) {
		t.Errorf("expected escaped parentheses in output")
	}
	if !bytes.Contains(out, []byte(`Level \\Six\\`)) {
		t.Errorf("expected escaped backslashes in output")
	}
}
