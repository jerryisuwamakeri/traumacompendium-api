package database

import (
	"strings"
	"testing"
)

func TestSplitStatements_SimpleStatements(t *testing.T) {
	got := splitStatements("SELECT 1; SELECT 2;")
	want := []string{"SELECT 1;", " SELECT 2;"}
	assertStatements(t, got, want)
}

func TestSplitStatements_SemicolonInsideStringLiteral_NotTreatedAsTerminator(t *testing.T) {
	// Regression test: a lesson seed migration contained the sentence
	// "...a flaw in you; it is a learned pattern..." inside a quoted
	// string value, which a naive strings.Split(script, ";") splits mid
	// statement, producing a dangling open quote and a MySQL syntax
	// error on the next INSERT run.
	script := `INSERT INTO lessons (title) VALUES ('That is not a flaw in you; it is a learned pattern.');`

	got := splitStatements(script)
	if len(got) != 1 {
		t.Fatalf("splitStatements() returned %d statements, want 1: %q", len(got), got)
	}
	if got[0] != script {
		t.Errorf("splitStatements()[0] = %q, want the statement unsplit: %q", got[0], script)
	}
}

func TestSplitStatements_ApostropheInsideLineComment_DoesNotOpenStringLiteral(t *testing.T) {
	// Regression test: a migration comment like "-- the brief's learning
	// framework" contains an apostrophe. A naive quote tracker treats
	// that as opening a string literal, desyncing every statement that
	// follows for the rest of the file.
	script := "-- the brief's learning framework\nINSERT INTO t (v) VALUES ('ok');"

	got := splitStatements(script)
	if len(got) != 1 {
		t.Fatalf("splitStatements() returned %d statements, want 1: %q", len(got), got)
	}
	if !strings.Contains(got[0], "INSERT INTO t (v) VALUES ('ok');") {
		t.Errorf("splitStatements()[0] = %q, want it to contain the full, unmangled INSERT", got[0])
	}
}

func TestSplitStatements_SemicolonInsideLineComment_NotTreatedAsTerminator(t *testing.T) {
	script := "-- One badge per level; awarded on completion\nINSERT INTO t (v) VALUES (1);"

	got := splitStatements(script)
	if len(got) != 1 {
		t.Fatalf("splitStatements() returned %d statements, want 1: %q", len(got), got)
	}
}

func TestSplitStatements_EscapedQuoteInsideStringLiteral(t *testing.T) {
	script := `INSERT INTO t (v) VALUES ('it''s fine; really');`

	got := splitStatements(script)
	if len(got) != 1 {
		t.Fatalf("splitStatements() returned %d statements, want 1: %q", len(got), got)
	}
}

func TestSplitStatements_MultipleStatementsWithStringSemicolons(t *testing.T) {
	script := "INSERT INTO t (v) VALUES ('a; b');\nINSERT INTO t (v) VALUES ('c; d');"

	got := splitStatements(script)
	if len(got) != 2 {
		t.Fatalf("splitStatements() returned %d statements, want 2: %q", len(got), got)
	}
}

func TestSplitStatements_IgnoresBlankTrailingSegment(t *testing.T) {
	got := splitStatements("SELECT 1;\n\n  \n")
	if len(got) != 1 {
		t.Fatalf("splitStatements() returned %d statements, want 1 (trailing whitespace dropped): %q", len(got), got)
	}
}

func assertStatements(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("splitStatements() returned %d statements, want %d: %q", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("statement[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
