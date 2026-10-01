package sms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

func TestNew_SelectsImplementationByMode(t *testing.T) {
	sandbox := New(config.TermiiConfig{Mode: config.ModeSandbox}, zerolog.Nop())
	if _, ok := sandbox.(*sandboxSender); !ok {
		t.Errorf("New(sandbox mode) = %T, want *sandboxSender", sandbox)
	}

	production := New(config.TermiiConfig{Mode: config.ModeProduction}, zerolog.Nop())
	if _, ok := production.(*termiiSender); !ok {
		t.Errorf("New(production mode) = %T, want *termiiSender", production)
	}
}

func TestSandboxSender_NeverErrors(t *testing.T) {
	sender := New(config.TermiiConfig{Mode: config.ModeSandbox}, zerolog.Nop())

	err := sender.Send(context.Background(), Message{To: "+2348012345678", Body: "Your code is 123456"})
	if err != nil {
		t.Errorf("sandboxSender.Send() error = %v, want nil", err)
	}
}

func TestBuildTermiiRequest(t *testing.T) {
	cfg := config.TermiiConfig{APIKey: "test-key", SenderID: "TraumaComp"}
	msg := Message{To: "+2348012345678", Body: "Your code is 123456"}

	got := buildTermiiRequest(cfg, msg)

	want := termiiSendRequest{
		APIKey:  "test-key",
		To:      "+2348012345678",
		From:    "TraumaComp",
		SMS:     "Your code is 123456",
		Type:    "plain",
		Channel: "generic",
	}
	if got != want {
		t.Errorf("buildTermiiRequest() = %+v, want %+v", got, want)
	}
}

func TestTermiiSender_Send_Success(t *testing.T) {
	var received termiiSendRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/sms/send" {
			t.Errorf("path = %s, want /api/sms/send", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(termiiSendResponse{
			MessageID: "msg-123",
			Message:   "Successfully Sent",
			Balance:   100.5,
			User:      "TraumaCompendium",
		})
	}))
	defer server.Close()

	sender := &termiiSender{
		cfg: config.TermiiConfig{
			BaseURL:  server.URL,
			APIKey:   "test-key",
			SenderID: "TraumaComp",
		},
		httpClient: server.Client(),
	}

	err := sender.Send(context.Background(), Message{To: "+2348012345678", Body: "Your code is 123456"})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if received.To != "+2348012345678" {
		t.Errorf("received.To = %q, want %q", received.To, "+2348012345678")
	}
	if received.APIKey != "test-key" {
		t.Errorf("received.APIKey = %q, want %q", received.APIKey, "test-key")
	}
}

func TestTermiiSender_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid API Key"}`))
	}))
	defer server.Close()

	sender := &termiiSender{
		cfg:        config.TermiiConfig{BaseURL: server.URL, APIKey: "bad-key"},
		httpClient: server.Client(),
	}

	err := sender.Send(context.Background(), Message{To: "+2348012345678", Body: "hi"})
	if err == nil {
		t.Fatal("Send() error = nil, want error for non-2xx response")
	}
}

func TestTermiiSender_Send_MissingMessageID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(termiiSendResponse{Message: "Something went wrong"})
	}))
	defer server.Close()

	sender := &termiiSender{
		cfg:        config.TermiiConfig{BaseURL: server.URL, APIKey: "test-key"},
		httpClient: server.Client(),
	}

	err := sender.Send(context.Background(), Message{To: "+2348012345678", Body: "hi"})
	if err == nil {
		t.Fatal("Send() error = nil, want error when termii omits message_id")
	}
}

func TestTermiiSender_Send_RespectsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &termiiSender{
		cfg:        config.TermiiConfig{BaseURL: server.URL, APIKey: "test-key"},
		httpClient: server.Client(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := sender.Send(ctx, Message{To: "+2348012345678", Body: "hi"})
	if err == nil {
		t.Fatal("Send() error = nil, want context deadline exceeded error")
	}
}
