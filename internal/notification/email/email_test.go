package email

import (
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

func TestNew_SelectsImplementationByMode(t *testing.T) {
	sandbox := New(config.SMTPConfig{Mode: config.ModeSandbox}, zerolog.Nop())
	if _, ok := sandbox.(*sandboxSender); !ok {
		t.Errorf("New(sandbox mode) = %T, want *sandboxSender", sandbox)
	}

	production := New(config.SMTPConfig{Mode: config.ModeProduction}, zerolog.Nop())
	if _, ok := production.(*smtpSender); !ok {
		t.Errorf("New(production mode) = %T, want *smtpSender", production)
	}
}

func TestSandboxSender_NeverErrors(t *testing.T) {
	sender := New(config.SMTPConfig{Mode: config.ModeSandbox}, zerolog.Nop())

	err := sender.Send(context.Background(), Message{
		To:      "user@example.com",
		Subject: "Verify your email",
		Body:    "Your code is 123456",
	})
	if err != nil {
		t.Errorf("sandboxSender.Send() error = %v, want nil", err)
	}
}

func TestBuildMIMEMessage(t *testing.T) {
	cfg := config.SMTPConfig{
		FromAddr: "no-reply@traumacompendium.app",
		FromName: "TraumaCompendium",
	}
	msg := Message{
		To:      "jane@example.com",
		Subject: "Your verification code",
		Body:    "Your code is 123456. It expires in 10 minutes.",
	}

	raw := string(buildMIMEMessage(cfg, msg))

	for _, want := range []string{
		"From: TraumaCompendium <no-reply@traumacompendium.app>",
		"To: jane@example.com",
		"Subject: Your verification code",
		"Your code is 123456. It expires in 10 minutes.",
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("buildMIMEMessage() missing %q in output:\n%s", want, raw)
		}
	}

	headerEnd := strings.Index(raw, "\r\n\r\n")
	if headerEnd == -1 {
		t.Fatal("buildMIMEMessage() output has no blank line separating headers from body")
	}
}
