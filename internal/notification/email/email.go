// Package email sends transactional email (OTP codes, password resets).
//
// Two implementations exist behind the same Sender interface: a sandbox
// sender that logs what would have been sent (used in development, tests,
// and CI so nothing ever dials out), and a real SMTP sender. New picks the
// right one from config.SMTPConfig.Mode, so callers never branch on mode
// themselves.
package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// Message is a plain-text transactional email.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender dispatches a single transactional email.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// New returns the Sender appropriate for cfg.Mode.
func New(cfg config.SMTPConfig, log zerolog.Logger) Sender {
	if cfg.Mode == config.ModeProduction {
		return &smtpSender{cfg: cfg}
	}
	return &sandboxSender{cfg: cfg, log: log}
}

// sandboxSender never performs network I/O. It exists so local
// development and CI can exercise the full registration/verification flow
// without an SMTP account, and so the OTP code is visible to the developer
// in the application log.
type sandboxSender struct {
	cfg config.SMTPConfig
	log zerolog.Logger
}

func (s *sandboxSender) Send(_ context.Context, msg Message) error {
	s.log.Info().
		Str("channel", "email").
		Str("mode", "sandbox").
		Str("to", msg.To).
		Str("subject", msg.Subject).
		Str("body", msg.Body).
		Msg("notification: email suppressed in sandbox mode (would have been sent)")
	return nil
}

// smtpSender sends real email via an SMTP relay using Go's standard
// library client — no third-party SDK required for plain SMTP+AUTH.
type smtpSender struct {
	cfg config.SMTPConfig
}

func (s *smtpSender) Send(ctx context.Context, msg Message) error {
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	body := buildMIMEMessage(s.cfg, msg)

	errCh := make(chan error, 1)
	go func() {
		errCh <- smtp.SendMail(addr, auth, s.cfg.FromAddr, []string{msg.To}, body)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email: send: %w", ctx.Err())
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("email: send: %w", err)
		}
		return nil
	}
}

// buildMIMEMessage renders a plain-text RFC 5322 message. It is a pure
// function (no I/O) so the header/body wiring can be unit tested without a
// real SMTP server.
func buildMIMEMessage(cfg config.SMTPConfig, msg Message) []byte {
	from := fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddr)

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", msg.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)

	return []byte(b.String())
}
