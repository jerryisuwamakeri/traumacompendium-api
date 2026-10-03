// Package pdf hand-rolls just enough of the PDF 1.4 spec to render a
// single-page certificate of completion: a header, a handful of
// indirect objects (catalog, page tree, one page, a content stream,
// two base-14 fonts), an xref table, and a trailer. There is no
// PDF-writing library in this project's offline module cache, and the
// certificate layout's needs are narrow enough (a few centered lines
// of text in Helvetica, a font every PDF viewer ships without
// embedding) that a general-purpose library would be overkill even if
// one were available.
package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// A4 landscape, in points (1/72 inch) — a natural shape for a
// certificate.
const (
	pageWidth  = 841.89
	pageHeight = 595.28
)

// NewCertificate renders a one-page certificate of completion
// naming the recipient, the level they completed, when, and a
// human-verifiable serial (see GET /certificates/verify/:serial).
// Text is restricted to printable ASCII (sanitizeText below) since
// this hand-rolled writer only carries the standard Helvetica metrics
// needed to center text, not a full Unicode-capable font.
func NewCertificate(name, levelName string, issuedAt time.Time, serial string) ([]byte, error) {
	content := buildContentStream(name, levelName, issuedAt, serial)

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")

	var offsets []int
	nextObj := func(body string) {
		offsets = append(offsets, buf.Len())
		buf.WriteString(body)
	}

	nextObj(fmt.Sprintf("%d 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n", len(offsets)+1))
	nextObj(fmt.Sprintf("%d 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n", len(offsets)+1))
	nextObj(fmt.Sprintf(
		"%d 0 obj\n<< /Type /Page /Parent 2 0 R /Resources << /Font << /F1 5 0 R /F2 6 0 R >> >> /MediaBox [0 0 %.2f %.2f] /Contents 4 0 R >>\nendobj\n",
		len(offsets)+1, pageWidth, pageHeight,
	))
	nextObj(fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(offsets)+1, len(content), content))
	nextObj(fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n", len(offsets)+1))
	nextObj(fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>\nendobj\n", len(offsets)+1))

	xrefStart := buf.Len()
	objCount := len(offsets) + 1 // +1 for the free-list head entry
	fmt.Fprintf(&buf, "xref\n0 %d\n", objCount)
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", objCount, xrefStart)

	return buf.Bytes(), nil
}

func buildContentStream(name, levelName string, issuedAt time.Time, serial string) string {
	name = sanitizeText(name)
	levelName = sanitizeText(levelName)

	var b strings.Builder
	b.WriteString("BT\n")
	centeredLine(&b, "F2", 28, pageHeight-150, "Certificate of Completion")
	centeredLine(&b, "F1", 14, pageHeight-210, "This certifies that")
	centeredLine(&b, "F2", 24, pageHeight-250, name)
	centeredLine(&b, "F1", 14, pageHeight-290, "has successfully completed")
	centeredLine(&b, "F2", 20, pageHeight-325, levelName)
	centeredLine(&b, "F1", 12, pageHeight-355, "on the Trauma Compendium learning path")
	centeredLine(&b, "F1", 11, 110, "Issued "+issuedAt.UTC().Format("January 2, 2006"))
	centeredLine(&b, "F1", 10, 85, "Certificate No. "+serial)
	b.WriteString("ET")

	return b.String()
}

func centeredLine(b *strings.Builder, font string, size, y float64, text string) {
	x := (pageWidth - textWidth(font, size, text)) / 2
	fmt.Fprintf(b, "/%s %g Tf\n1 0 0 1 %.2f %.2f Tm\n(%s) Tj\n", font, size, x, y, escapeText(text))
}

// escapeText escapes the three characters that are meaningful inside a
// PDF literal string: backslash and the two parentheses.
func escapeText(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	return s
}

// sanitizeText replaces anything outside printable ASCII with '?' —
// the Helvetica width table below only covers that range, and this
// writer uses the font's built-in WinAnsi-ish encoding rather than
// embedding a custom one.
func sanitizeText(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r < 0x20 || r > 0x7E {
			r = '?'
		}
		out = append(out, r)
	}
	return string(out)
}

func textWidth(font string, size float64, text string) float64 {
	table := helveticaWidths
	if font == "F2" {
		table = helveticaBoldWidths
	}
	var total int
	for _, r := range text {
		if w, ok := table[r]; ok {
			total += w
		} else {
			total += 556
		}
	}
	return float64(total) / 1000 * size
}

// Standard Adobe AFM metrics (widths per 1000 units of em) for the
// base-14 Helvetica and Helvetica-Bold fonts, printable ASCII only —
// public, part of the PDF spec itself, and needed here only to center
// text reasonably rather than guess at an average character width.
var helveticaWidths = map[rune]int{
	' ': 278, '!': 278, '"': 355, '#': 556, '$': 556, '%': 889, '&': 667, '\'': 191,
	'(': 333, ')': 333, '*': 389, '+': 584, ',': 278, '-': 333, '.': 278, '/': 278,
	'0': 556, '1': 556, '2': 556, '3': 556, '4': 556, '5': 556, '6': 556, '7': 556, '8': 556, '9': 556,
	':': 278, ';': 278, '<': 584, '=': 584, '>': 584, '?': 556, '@': 1015,
	'A': 667, 'B': 667, 'C': 722, 'D': 722, 'E': 667, 'F': 611, 'G': 778, 'H': 722, 'I': 278, 'J': 500,
	'K': 667, 'L': 556, 'M': 833, 'N': 722, 'O': 778, 'P': 667, 'Q': 778, 'R': 722, 'S': 667, 'T': 611,
	'U': 722, 'V': 667, 'W': 944, 'X': 667, 'Y': 667, 'Z': 611,
	'[': 278, '\\': 278, ']': 278, '^': 469, '_': 556, '`': 333,
	'a': 556, 'b': 556, 'c': 500, 'd': 556, 'e': 556, 'f': 278, 'g': 556, 'h': 556, 'i': 222, 'j': 222,
	'k': 500, 'l': 222, 'm': 833, 'n': 556, 'o': 556, 'p': 556, 'q': 556, 'r': 333, 's': 500, 't': 278,
	'u': 556, 'v': 500, 'w': 722, 'x': 500, 'y': 500, 'z': 500,
	'{': 334, '|': 260, '}': 334, '~': 584,
}

var helveticaBoldWidths = map[rune]int{
	' ': 278, '!': 333, '"': 474, '#': 556, '$': 556, '%': 889, '&': 722, '\'': 238,
	'(': 333, ')': 333, '*': 389, '+': 584, ',': 278, '-': 333, '.': 278, '/': 278,
	'0': 556, '1': 556, '2': 556, '3': 556, '4': 556, '5': 556, '6': 556, '7': 556, '8': 556, '9': 556,
	':': 333, ';': 333, '<': 584, '=': 584, '>': 584, '?': 611, '@': 975,
	'A': 722, 'B': 722, 'C': 722, 'D': 722, 'E': 667, 'F': 611, 'G': 778, 'H': 722, 'I': 278, 'J': 556,
	'K': 722, 'L': 611, 'M': 833, 'N': 722, 'O': 778, 'P': 667, 'Q': 778, 'R': 722, 'S': 667, 'T': 611,
	'U': 722, 'V': 667, 'W': 944, 'X': 667, 'Y': 667, 'Z': 611,
	'[': 333, '\\': 278, ']': 333, '^': 584, '_': 556, '`': 333,
	'a': 556, 'b': 611, 'c': 556, 'd': 611, 'e': 556, 'f': 333, 'g': 611, 'h': 611, 'i': 278, 'j': 278,
	'k': 556, 'l': 278, 'm': 889, 'n': 611, 'o': 611, 'p': 611, 'q': 611, 'r': 389, 's': 556, 't': 333,
	'u': 611, 'v': 556, 'w': 778, 'x': 556, 'y': 556, 'z': 500,
	'{': 389, '|': 280, '}': 389, '~': 584,
}
