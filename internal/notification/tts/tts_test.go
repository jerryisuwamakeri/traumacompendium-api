package tts

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

func TestNew_SelectsImplementationByMode(t *testing.T) {
	sandbox := New(config.TTSConfig{Mode: config.ModeSandbox}, zerolog.Nop())
	if _, ok := sandbox.(*sandboxSynthesizer); !ok {
		t.Errorf("New(sandbox mode) = %T, want *sandboxSynthesizer", sandbox)
	}

	production := New(config.TTSConfig{Mode: config.ModeProduction}, zerolog.Nop())
	if _, ok := production.(*googleCloudSynthesizer); !ok {
		t.Errorf("New(production mode) = %T, want *googleCloudSynthesizer", production)
	}
}

func TestSandboxSynthesizer_ReturnsValidPlaceholderAudio(t *testing.T) {
	synth := New(config.TTSConfig{Mode: config.ModeSandbox}, zerolog.Nop())

	result, err := synth.Synthesize(context.Background(), Request{Text: "Hello", Locale: "yo"})
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}
	if result.ContentType != "audio/wav" {
		t.Errorf("ContentType = %q, want audio/wav", result.ContentType)
	}
	if len(result.Audio) == 0 {
		t.Error("expected non-empty placeholder audio")
	}
	// A valid WAV starts with "RIFF" and declares "WAVE" shortly after.
	if !bytes.HasPrefix(result.Audio, []byte("RIFF")) {
		t.Error("placeholder audio is not a valid WAV (missing RIFF header)")
	}
	if !bytes.Contains(result.Audio[:12], []byte("WAVE")) {
		t.Error("placeholder audio is not a valid WAV (missing WAVE marker)")
	}
}

func TestGoogleCloudSynthesizer_Synthesize_Success(t *testing.T) {
	wantAudio := []byte("fake-mp3-bytes")
	var received synthesizeRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/text:synthesize" {
			t.Errorf("path = %s, want /v1/text:synthesize", r.URL.Path)
		}
		if r.URL.Query().Get("key") != "test-api-key" {
			t.Errorf("key query param = %q, want test-api-key", r.URL.Query().Get("key"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(synthesizeResponse{
			AudioContent: base64.StdEncoding.EncodeToString(wantAudio),
		})
	}))
	defer server.Close()

	synth := &googleCloudSynthesizer{
		cfg:        config.TTSConfig{BaseURL: server.URL, APIKey: "test-api-key"},
		httpClient: server.Client(),
	}

	result, err := synth.Synthesize(context.Background(), Request{Text: "Bonjour", Locale: "fr"})
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}
	if !bytes.Equal(result.Audio, wantAudio) {
		t.Errorf("Audio = %q, want %q", result.Audio, wantAudio)
	}
	if result.ContentType != "audio/mpeg" {
		t.Errorf("ContentType = %q, want audio/mpeg", result.ContentType)
	}
	if received.Voice.LanguageCode != "fr-FR" {
		t.Errorf("request languageCode = %q, want fr-FR", received.Voice.LanguageCode)
	}
	if received.Input.Text != "Bonjour" {
		t.Errorf("request text = %q, want Bonjour", received.Input.Text)
	}
}

func TestGoogleCloudSynthesizer_UnmappedLocale_ReturnsErrorWithoutCallingProvider(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	synth := &googleCloudSynthesizer{
		cfg:        config.TTSConfig{BaseURL: server.URL, APIKey: "test-api-key"},
		httpClient: server.Client(),
	}

	_, err := synth.Synthesize(context.Background(), Request{Text: "hi", Locale: "zz"})
	if err == nil {
		t.Fatal("Synthesize() with an unmapped locale error = nil, want error")
	}
	if called {
		t.Error("provider was called despite no voice mapping for the locale")
	}
}

func TestGoogleCloudSynthesizer_NonOKStatus_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"API key not valid"}}`))
	}))
	defer server.Close()

	synth := &googleCloudSynthesizer{
		cfg:        config.TTSConfig{BaseURL: server.URL, APIKey: "bad-key"},
		httpClient: server.Client(),
	}

	_, err := synth.Synthesize(context.Background(), Request{Text: "hi", Locale: "en"})
	if err == nil {
		t.Fatal("Synthesize() error = nil, want error for non-2xx response")
	}
}
