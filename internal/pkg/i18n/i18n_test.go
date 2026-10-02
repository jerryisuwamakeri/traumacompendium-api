package i18n

import "testing"

func TestT_KnownLocaleAndKey(t *testing.T) {
	got := T("fr", KeyRequired)
	want := "est requis"
	if got != want {
		t.Errorf("T(fr, KeyRequired) = %q, want %q", got, want)
	}
}

func TestT_UnsupportedLocale_FallsBackToEnglish(t *testing.T) {
	got := T("zz", KeyRequired)
	want := catalog["en"][KeyRequired]
	if got != want {
		t.Errorf("T(zz, KeyRequired) = %q, want English fallback %q", got, want)
	}
}

func TestT_EveryLocaleHasEveryKey(t *testing.T) {
	keys := []Key{KeyRequired, KeyEmail, KeyTooShort, KeyTooLong, KeyInvalidPhone, KeyInvalid}
	for locale, messages := range catalog {
		for _, k := range keys {
			if messages[k] == "" {
				t.Errorf("catalog[%q] is missing a message for key %q", locale, k)
			}
		}
	}
}
