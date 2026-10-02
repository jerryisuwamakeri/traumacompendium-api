// Package sanitize strips HTML markup from fields that are meant to be
// plain text — journal entries, mood notes, display names — so that if a
// frontend ever renders user-submitted text as HTML without escaping it
// (stored XSS), there's nothing left to execute by the time it reaches
// storage.
//
// This is defense in depth, not the primary defense: this API is
// JSON-only (Go's encoder already escapes `<`/`>`/`&` on the way out)
// and all SQL is parameterized, so there is no first-order XSS hole in
// this service itself. The risk this guards against is entirely in a
// future consumer's rendering code.
package sanitize

import (
	"strings"

	"golang.org/x/net/html"
)

// StripHTML removes all HTML tags, comments, and doctypes from input,
// keeping only the text content — plain prose with no markup passes
// through unchanged (including characters that merely look like the
// start of a tag, e.g. "I felt <3 about it", which the HTML5
// tokenization algorithm itself treats as literal text, not a tag).
//
// The text content of <script> and <style> elements is dropped along
// with their tags, since that content was never meant to be visible
// prose in the first place — everything else's inner text is kept (e.g.
// "<b>bold</b>" becomes "bold").
func StripHTML(input string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(input))
	var out strings.Builder
	skipDepth := 0

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break // end of input (io.EOF) or an unrecoverable parse error — either way, stop
		}

		switch tt {
		case html.StartTagToken:
			tok := tokenizer.Token()
			if isSkippedElement(tok.Data) {
				skipDepth++
			}
		case html.SelfClosingTagToken:
			// no text content to worry about (e.g. <br/>), and it never
			// opens a skip region
		case html.EndTagToken:
			tok := tokenizer.Token()
			if isSkippedElement(tok.Data) && skipDepth > 0 {
				skipDepth--
			}
		case html.TextToken:
			if skipDepth == 0 {
				out.Write(tokenizer.Text())
			}
		}
	}

	return out.String()
}

func isSkippedElement(tagName string) bool {
	return tagName == "script" || tagName == "style"
}
