package translate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

func TestNew_SelectsImplementationByMode(t *testing.T) {
	sandbox := New(config.TranslateConfig{Mode: config.ModeSandbox}, zerolog.Nop())
	if _, ok := sandbox.(*sandboxTranslator); !ok {
		t.Errorf("New(sandbox mode) = %T, want *sandboxTranslator", sandbox)
	}

	production := New(config.TranslateConfig{Mode: config.ModeProduction}, zerolog.Nop())
	if _, ok := production.(*googleCloudTranslator); !ok {
		t.Errorf("New(production mode) = %T, want *googleCloudTranslator", production)
	}
}

func TestSandboxTranslator_ReturnsTaggedPlaceholder_WithoutCallingOut(t *testing.T) {
	tr := New(config.TranslateConfig{Mode: config.ModeSandbox}, zerolog.Nop())

	result, err := tr.Translate(context.Background(), Request{
		Text:         "How are you feeling today?",
		SourceLocale: "en",
		TargetLocale: "yo",
	})
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}
	if !strings.Contains(result.TranslatedText, "yo") {
		t.Errorf("TranslatedText = %q, want it to be tagged with the target locale", result.TranslatedText)
	}
	if !strings.Contains(result.TranslatedText, "How are you feeling today?") {
		t.Errorf("TranslatedText = %q, want it to contain the original text", result.TranslatedText)
	}
}

func TestGoogleCloudTranslator_Translate_Success(t *testing.T) {
	var received translateRequestBody

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/language/translate2" {
			t.Errorf("path = %s, want /language/translate2", r.URL.Path)
		}
		if r.URL.Query().Get("key") != "test-api-key" {
			t.Errorf("key query param = %q, want test-api-key", r.URL.Query().Get("key"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"translations": []map[string]string{
					{"translatedText": "Bonjour le monde"},
				},
			},
		})
	}))
	defer server.Close()

	tr := &googleCloudTranslator{
		cfg:        config.TranslateConfig{BaseURL: server.URL, APIKey: "test-api-key"},
		httpClient: server.Client(),
	}

	result, err := tr.Translate(context.Background(), Request{
		Text:         "Hello world",
		SourceLocale: "en",
		TargetLocale: "fr",
	})
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}
	if result.TranslatedText != "Bonjour le monde" {
		t.Errorf("TranslatedText = %q, want %q", result.TranslatedText, "Bonjour le monde")
	}
	if received.Target != "fr" {
		t.Errorf("request target = %q, want fr", received.Target)
	}
	if received.Source != "en" {
		t.Errorf("request source = %q, want en", received.Source)
	}
	if len(received.Q) != 1 || received.Q[0] != "Hello world" {
		t.Errorf("request q = %v, want [Hello world]", received.Q)
	}
}

func TestGoogleCloudTranslator_NonOKStatus_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"API key not valid"}}`))
	}))
	defer server.Close()

	tr := &googleCloudTranslator{
		cfg:        config.TranslateConfig{BaseURL: server.URL, APIKey: "bad-key"},
		httpClient: server.Client(),
	}

	_, err := tr.Translate(context.Background(), Request{Text: "hi", SourceLocale: "en", TargetLocale: "yo"})
	if err == nil {
		t.Fatal("Translate() error = nil, want error for non-2xx response")
	}
}

func TestGoogleCloudTranslator_NoTranslationsInResponse_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"translations": []map[string]string{}},
		})
	}))
	defer server.Close()

	tr := &googleCloudTranslator{
		cfg:        config.TranslateConfig{BaseURL: server.URL, APIKey: "test-api-key"},
		httpClient: server.Client(),
	}

	_, err := tr.Translate(context.Background(), Request{Text: "hi", SourceLocale: "en", TargetLocale: "yo"})
	if err == nil {
		t.Fatal("Translate() error = nil, want error when the provider returns zero translations")
	}
}
