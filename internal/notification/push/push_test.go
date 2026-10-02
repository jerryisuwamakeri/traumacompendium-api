package push

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
	sandbox := New(config.PushConfig{Mode: config.ModeSandbox}, zerolog.Nop())
	if _, ok := sandbox.(*sandboxSender); !ok {
		t.Errorf("New(sandbox mode) = %T, want *sandboxSender", sandbox)
	}

	production := New(config.PushConfig{Mode: config.ModeProduction}, zerolog.Nop())
	if _, ok := production.(*fcmSender); !ok {
		t.Errorf("New(production mode) = %T, want *fcmSender", production)
	}
}

func TestSandboxSender_NeverErrors(t *testing.T) {
	sender := New(config.PushConfig{Mode: config.ModeSandbox}, zerolog.Nop())

	err := sender.Send(context.Background(), Message{DeviceToken: "device-token", Title: "Alert", Body: "Something happened"})
	if err != nil {
		t.Errorf("sandboxSender.Send() error = %v, want nil", err)
	}
}

func TestFCMSender_Send_Success(t *testing.T) {
	var received fcmSendRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/fcm/send" {
			t.Errorf("path = %s, want /fcm/send", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "key=test-server-key" {
			t.Errorf("Authorization = %q, want key=test-server-key", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fcmSendResponse{Success: 1, Failure: 0})
	}))
	defer server.Close()

	sender := &fcmSender{
		cfg:        config.PushConfig{BaseURL: server.URL, ServerKey: "test-server-key"},
		httpClient: server.Client(),
	}

	err := sender.Send(context.Background(), Message{
		DeviceToken: "device-token", Title: "SafeReach Alert", Body: "Jane triggered an urgent alert.",
		Data: map[string]string{"alert_id": "abc-123"},
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if received.To != "device-token" {
		t.Errorf("received.To = %q, want device-token", received.To)
	}
	if received.Notification.Title != "SafeReach Alert" {
		t.Errorf("received.Notification.Title = %q, want SafeReach Alert", received.Notification.Title)
	}
	if received.Data["alert_id"] != "abc-123" {
		t.Errorf("received.Data[alert_id] = %q, want abc-123", received.Data["alert_id"])
	}
}

func TestFCMSender_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"InvalidServerKey"}`))
	}))
	defer server.Close()

	sender := &fcmSender{
		cfg:        config.PushConfig{BaseURL: server.URL, ServerKey: "bad-key"},
		httpClient: server.Client(),
	}

	err := sender.Send(context.Background(), Message{DeviceToken: "device-token", Title: "t", Body: "b"})
	if err == nil {
		t.Fatal("Send() error = nil, want error for non-2xx response")
	}
}

func TestFCMSender_Send_ReportsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fcmSendResponse{Success: 0, Failure: 1})
	}))
	defer server.Close()

	sender := &fcmSender{
		cfg:        config.PushConfig{BaseURL: server.URL, ServerKey: "test-server-key"},
		httpClient: server.Client(),
	}

	err := sender.Send(context.Background(), Message{DeviceToken: "stale-token", Title: "t", Body: "b"})
	if err == nil {
		t.Fatal("Send() error = nil, want error when fcm reports a failure (e.g. invalid/unregistered token)")
	}
}

func TestFCMSender_Send_RespectsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &fcmSender{
		cfg:        config.PushConfig{BaseURL: server.URL, ServerKey: "test-server-key"},
		httpClient: server.Client(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := sender.Send(ctx, Message{DeviceToken: "device-token", Title: "t", Body: "b"})
	if err == nil {
		t.Fatal("Send() error = nil, want context deadline exceeded error")
	}
}
