package hash

import "testing"

func TestPassword_RoundTrip(t *testing.T) {
	hashed, err := Password("correct horse battery staple")
	if err != nil {
		t.Fatalf("Password() error = %v", err)
	}

	if !PasswordMatches(hashed, "correct horse battery staple") {
		t.Error("PasswordMatches() = false, want true for correct password")
	}

	if PasswordMatches(hashed, "wrong password") {
		t.Error("PasswordMatches() = true, want false for wrong password")
	}
}

func TestPassword_DistinctHashesForSameInput(t *testing.T) {
	h1, err := Password("same-input")
	if err != nil {
		t.Fatalf("Password() error = %v", err)
	}
	h2, err := Password("same-input")
	if err != nil {
		t.Fatalf("Password() error = %v", err)
	}

	if h1 == h2 {
		t.Error("expected distinct hashes due to bcrypt salting, got identical hashes")
	}
}

func TestToken_RoundTrip(t *testing.T) {
	digest := Token("123456")

	if !TokenMatches(digest, "123456") {
		t.Error("TokenMatches() = false, want true for matching token")
	}

	if TokenMatches(digest, "654321") {
		t.Error("TokenMatches() = true, want false for non-matching token")
	}
}

func TestToken_Deterministic(t *testing.T) {
	first := Token("abc")
	second := Token("abc")
	if first != second {
		t.Errorf("Token() = %q then %q, want the same digest for the same input", first, second)
	}
}
