package fakes

import (
	"context"
	"errors"
	"sync"

	"github.com/traumacompendium/backend/internal/notification/email"
	"github.com/traumacompendium/backend/internal/notification/push"
	"github.com/traumacompendium/backend/internal/notification/sms"
)

// EmailSender records every message it's asked to send, so tests can
// assert on delivery (recipient, subject, body) without a real SMTP
// transport.
type EmailSender struct {
	mu   sync.Mutex
	Sent []email.Message
}

func (f *EmailSender) Send(_ context.Context, msg email.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Sent = append(f.Sent, msg)
	return nil
}

func (f *EmailSender) Last() email.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Sent[len(f.Sent)-1]
}

// SMSSender mirrors EmailSender for the sms.Sender interface.
type SMSSender struct {
	mu   sync.Mutex
	Sent []sms.Message
}

func (f *SMSSender) Send(_ context.Context, msg sms.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Sent = append(f.Sent, msg)
	return nil
}

func (f *SMSSender) Last() sms.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Sent[len(f.Sent)-1]
}

// PushSender mirrors EmailSender for the push.Sender interface.
type PushSender struct {
	mu   sync.Mutex
	Sent []push.Message
}

func (f *PushSender) Send(_ context.Context, msg push.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Sent = append(f.Sent, msg)
	return nil
}

func (f *PushSender) Last() push.Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Sent[len(f.Sent)-1]
}

// FailingEmailSender always fails — used to assert that OTP delivery
// failures don't abort the flow that triggered them.
type FailingEmailSender struct{}

func (FailingEmailSender) Send(context.Context, email.Message) error {
	return errSimulatedDeliveryFailure
}

var errSimulatedDeliveryFailure = errors.New("fakes: simulated delivery failure")
