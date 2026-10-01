package otp

import (
	"testing"
	"unicode"
)

func TestGenerate_Length(t *testing.T) {
	for _, length := range []int{4, 6, 8} {
		code, err := Generate(length)
		if err != nil {
			t.Fatalf("Generate(%d) error = %v", length, err)
		}
		if len(code) != length {
			t.Errorf("Generate(%d) = %q, want length %d", length, code, length)
		}
		for _, r := range code {
			if !unicode.IsDigit(r) {
				t.Errorf("Generate(%d) = %q, contains non-digit %q", length, code, r)
			}
		}
	}
}

func TestGenerate_RejectsNonPositiveLength(t *testing.T) {
	if _, err := Generate(0); err == nil {
		t.Error("Generate(0) error = nil, want error")
	}
	if _, err := Generate(-1); err == nil {
		t.Error("Generate(-1) error = nil, want error")
	}
}

func TestGenerate_IsNotConstant(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		code, err := Generate(6)
		if err != nil {
			t.Fatalf("Generate(6) error = %v", err)
		}
		seen[code] = true
	}

	if len(seen) < 2 {
		t.Error("Generate(6) produced the same code across 20 calls, expected randomness")
	}
}
