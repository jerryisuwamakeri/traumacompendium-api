package sanitize

import "testing"

func TestStripHTML_PlainTextUnchanged(t *testing.T) {
	in := "Today was hard. I'm proud of how I handled it, though."
	if got := StripHTML(in); got != in {
		t.Errorf("StripHTML(plain text) = %q, want unchanged %q", got, in)
	}
}

func TestStripHTML_CharacterThatLooksLikeATag_TreatedAsText(t *testing.T) {
	// "<3" is not a valid HTML tag start (tag names must start with a
	// letter) — the HTML5 tokenization algorithm (which the tokenizer
	// implements) treats this as literal text, not markup.
	in := "I felt <3 about it today"
	if got := StripHTML(in); got != in {
		t.Errorf("StripHTML(%q) = %q, want unchanged (not a real tag)", in, got)
	}
}

func TestStripHTML_RemovesScriptTagAndItsContent(t *testing.T) {
	in := `Hello<script>alert(document.cookie)</script>World`
	want := "HelloWorld"
	if got := StripHTML(in); got != want {
		t.Errorf("StripHTML(%q) = %q, want %q", in, got, want)
	}
}

func TestStripHTML_RemovesStyleTagAndItsContent(t *testing.T) {
	in := `Note<style>body{display:none}</style>Here`
	want := "NoteHere"
	if got := StripHTML(in); got != want {
		t.Errorf("StripHTML(%q) = %q, want %q", in, got, want)
	}
}

func TestStripHTML_KeepsTextContentOfOtherTags(t *testing.T) {
	in := "<b>Bold</b> and <i>italic</i> text"
	want := "Bold and italic text"
	if got := StripHTML(in); got != want {
		t.Errorf("StripHTML(%q) = %q, want %q", in, got, want)
	}
}

func TestStripHTML_RemovesImgOnErrorPayload(t *testing.T) {
	in := `<img src=x onerror="alert(1)">Caption`
	want := "Caption"
	if got := StripHTML(in); got != want {
		t.Errorf("StripHTML(%q) = %q, want %q", in, got, want)
	}
}

func TestStripHTML_EmptyInput(t *testing.T) {
	if got := StripHTML(""); got != "" {
		t.Errorf("StripHTML(\"\") = %q, want empty", got)
	}
}
