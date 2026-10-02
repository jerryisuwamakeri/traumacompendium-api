// Package translate turns canonical English content into a draft
// translation for another locale, mirroring the tts package's shape
// exactly: a Translator interface, a sandbox implementation (used in
// dev/CI/tests — no real provider call), and a production implementation,
// selected by config.TranslateConfig.Mode.
//
// This is never wired into the live HTTP API — translating static
// lesson/quiz/badge content is a content-pipeline action performed ahead
// of time by cmd/translate-content, the same reasoning as why lesson
// text is pre-translated rather than translated on every request.
//
// Critically, a translation produced here — whichever Translator
// produced it — is still a machine draft, not a human-reviewed one.
// Callers must keep marking these rows needs_review: true regardless of
// source; this package only changes which machine produced the draft.
package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// Request is a single piece of canonical-English text to translate into
// TargetLocale (one of our locale codes, e.g. "yo", "fr").
type Request struct {
	Text         string
	SourceLocale string
	TargetLocale string
}

// Result is a translated piece of text.
type Result struct {
	TranslatedText string
}

// Translator turns text from one locale into another.
type Translator interface {
	Translate(ctx context.Context, req Request) (Result, error)
}

// New returns the Translator appropriate for cfg.Mode.
func New(cfg config.TranslateConfig, log zerolog.Logger) Translator {
	if cfg.Mode == config.ModeProduction {
		return &googleCloudTranslator{
			cfg:        cfg,
			httpClient: &http.Client{Timeout: 30 * time.Second},
		}
	}
	return &sandboxTranslator{log: log}
}

// sandboxTranslator never calls a real provider. It returns a tagged
// placeholder — the target locale prefixed onto the original text — so
// callers exercising the full "translate, upsert, serve" pipeline in
// dev/CI get a real, distinguishable string end to end, just not a real
// translation.
type sandboxTranslator struct {
	log zerolog.Logger
}

func (s *sandboxTranslator) Translate(_ context.Context, req Request) (Result, error) {
	s.log.Info().
		Str("channel", "translate").
		Str("provider", "google-cloud").
		Str("mode", "sandbox").
		Str("source_locale", req.SourceLocale).
		Str("target_locale", req.TargetLocale).
		Int("text_length", len(req.Text)).
		Msg("notification: translation suppressed in sandbox mode (would have been translated)")
	return Result{TranslatedText: fmt.Sprintf("[%s] %s", req.TargetLocale, req.Text)}, nil
}

// googleCloudTranslator calls Google Cloud Translation API v2.
// https://cloud.google.com/translate/docs/reference/rest/v2/translate
//
// Our locale codes (yo, ha, ig, fr, sw, ar) are passed directly as the
// target language — unlike tts's BCP-47 voice codes, Translation API v2
// just wants the plain ISO code, which is what SupportedLocales already
// uses. Whether the live API actually supports every one of our locales
// (pcm, in particular, is not a standard ISO 639-1 code) is not something
// this sandbox can verify; an unsupported target fails clearly via the
// API's own error response rather than silently mistranslating.
type googleCloudTranslator struct {
	cfg        config.TranslateConfig
	httpClient *http.Client
}

type translateRequestBody struct {
	Q      []string `json:"q"`
	Source string   `json:"source"`
	Target string   `json:"target"`
	Format string   `json:"format"`
}

type translateResponseBody struct {
	Data struct {
		Translations []struct {
			TranslatedText string `json:"translatedText"`
		} `json:"translations"`
	} `json:"data"`
}

func (g *googleCloudTranslator) Translate(ctx context.Context, req Request) (Result, error) {
	payload, err := json.Marshal(translateRequestBody{
		Q:      []string{req.Text},
		Source: req.SourceLocale,
		Target: req.TargetLocale,
		Format: "text",
	})
	if err != nil {
		return Result{}, fmt.Errorf("translate: encode request: %w", err)
	}

	endpoint := g.cfg.BaseURL + "/language/translate2?key=" + g.cfg.APIKey
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return Result{}, fmt.Errorf("translate: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return Result{}, fmt.Errorf("translate: request google cloud: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response already fully read below

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("translate: read google cloud response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("translate: google cloud returned %d for target %q: %s", resp.StatusCode, req.TargetLocale, string(body))
	}

	var parsed translateResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Result{}, fmt.Errorf("translate: decode google cloud response: %w", err)
	}

	if len(parsed.Data.Translations) == 0 {
		return Result{}, fmt.Errorf("translate: google cloud returned no translations for target %q", req.TargetLocale)
	}

	return Result{TranslatedText: parsed.Data.Translations[0].TranslatedText}, nil
}
