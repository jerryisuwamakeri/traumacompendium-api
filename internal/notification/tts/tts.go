// Package tts synthesizes lesson content into speech, mirroring the
// email/sms packages' shape exactly: a Sender-style interface with a
// sandbox implementation (used in dev/CI/tests — no real provider call)
// and a production implementation, selected by config.TTSConfig.Mode.
//
// Which languages are actually synthesizable depends entirely on what
// the live provider supports at request time — this package does not
// claim Google Cloud TTS supports all of learning.SupportedLocales, only
// that it will try the best-effort BCP-47 mapping in localeVoice and
// surface a clear error (via Synthesize's returned error) if the
// provider rejects a given locale.
package tts

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
)

// Request is a single piece of text to synthesize in a given locale.
type Request struct {
	Text   string
	Locale string
}

// Result is a synthesized audio clip.
type Result struct {
	Audio       []byte
	ContentType string
}

// Synthesizer turns text into speech.
type Synthesizer interface {
	Synthesize(ctx context.Context, req Request) (Result, error)
}

// New returns the Synthesizer appropriate for cfg.Mode.
func New(cfg config.TTSConfig, log zerolog.Logger) Synthesizer {
	if cfg.Mode == config.ModeProduction {
		return &googleCloudSynthesizer{
			cfg:        cfg,
			httpClient: &http.Client{Timeout: 30 * time.Second},
		}
	}
	return &sandboxSynthesizer{log: log}
}

// localeVoice maps our locale codes to a best-effort Google Cloud TTS
// BCP-47 language code. Not every entry here is guaranteed to be a voice
// Google actually ships — see the package doc comment.
var localeVoice = map[string]string{
	"en":  "en-US",
	"pcm": "pcm-NG",
	"yo":  "yo-NG",
	"ha":  "ha-NG",
	"ig":  "ig-NG",
	"fr":  "fr-FR",
	"sw":  "sw-KE",
	"ar":  "ar-XA",
}

// sandboxSynthesizer never calls a real provider. It returns a short,
// valid, silent WAV clip (not an empty byte slice) so callers exercising
// the full "synthesize, cache, serve" path in dev/CI get a real audio
// file end to end, just a silent one.
type sandboxSynthesizer struct {
	log zerolog.Logger
}

func (s *sandboxSynthesizer) Synthesize(_ context.Context, req Request) (Result, error) {
	s.log.Info().
		Str("channel", "tts").
		Str("provider", "google-cloud").
		Str("mode", "sandbox").
		Str("locale", req.Locale).
		Int("text_length", len(req.Text)).
		Msg("notification: tts synthesis suppressed in sandbox mode (would have been synthesized)")
	return Result{Audio: silentWAV(), ContentType: "audio/wav"}, nil
}

// googleCloudSynthesizer calls Google Cloud Text-to-Speech's REST API.
// https://cloud.google.com/text-to-speech/docs/reference/rest/v1/text/synthesize
type googleCloudSynthesizer struct {
	cfg        config.TTSConfig
	httpClient *http.Client
}

type synthesizeRequest struct {
	Input       synthesizeInput       `json:"input"`
	Voice       synthesizeVoice       `json:"voice"`
	AudioConfig synthesizeAudioConfig `json:"audioConfig"`
}

type synthesizeInput struct {
	Text string `json:"text"`
}

type synthesizeVoice struct {
	LanguageCode string `json:"languageCode"`
	SSMLGender   string `json:"ssmlGender"`
}

type synthesizeAudioConfig struct {
	AudioEncoding string `json:"audioEncoding"`
}

type synthesizeResponse struct {
	AudioContent string `json:"audioContent"` // base64-encoded
}

func (s *googleCloudSynthesizer) Synthesize(ctx context.Context, req Request) (Result, error) {
	languageCode, ok := localeVoice[req.Locale]
	if !ok {
		return Result{}, fmt.Errorf("tts: no voice mapping configured for locale %q", req.Locale)
	}

	payload, err := json.Marshal(synthesizeRequest{
		Input:       synthesizeInput{Text: req.Text},
		Voice:       synthesizeVoice{LanguageCode: languageCode, SSMLGender: "NEUTRAL"},
		AudioConfig: synthesizeAudioConfig{AudioEncoding: "MP3"},
	})
	if err != nil {
		return Result{}, fmt.Errorf("tts: encode request: %w", err)
	}

	endpoint := s.cfg.BaseURL + "/v1/text:synthesize?key=" + s.cfg.APIKey
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return Result{}, fmt.Errorf("tts: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return Result{}, fmt.Errorf("tts: request google cloud: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response already fully read below

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("tts: read google cloud response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("tts: google cloud returned %d for locale %q: %s", resp.StatusCode, req.Locale, string(body))
	}

	var parsed synthesizeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Result{}, fmt.Errorf("tts: decode google cloud response: %w", err)
	}

	audio, err := base64.StdEncoding.DecodeString(parsed.AudioContent)
	if err != nil {
		return Result{}, fmt.Errorf("tts: decode base64 audio: %w", err)
	}

	return Result{Audio: audio, ContentType: "audio/mpeg"}, nil
}

// silentWAV is a minimal, valid, ~0.1s silent 8kHz mono WAV file — small
// enough to inline, real enough that anything parsing/serving it as
// audio (Content-Type, byte-length headers, an <audio> tag) behaves
// exactly as it would for real synthesized speech.
func silentWAV() []byte {
	const sampleRate = 8000
	const numSamples = 800 // 0.1s
	dataSize := numSamples * 2
	buf := new(bytes.Buffer)

	buf.WriteString("RIFF")
	writeUint32LE(buf, uint32(36+dataSize)) //nolint:gosec // dataSize is a small compile-time-bounded constant
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	writeUint32LE(buf, 16)
	writeUint16LE(buf, 1) // PCM
	writeUint16LE(buf, 1) // mono
	writeUint32LE(buf, sampleRate)
	writeUint32LE(buf, sampleRate*2) // byte rate
	writeUint16LE(buf, 2)            // block align
	writeUint16LE(buf, 16)           // bits per sample
	buf.WriteString("data")
	writeUint32LE(buf, uint32(dataSize)) //nolint:gosec // dataSize is a small compile-time-bounded constant
	buf.Write(make([]byte, dataSize))    // silence

	return buf.Bytes()
}

func writeUint32LE(buf *bytes.Buffer, v uint32) {
	buf.WriteByte(byte(v))
	buf.WriteByte(byte(v >> 8))
	buf.WriteByte(byte(v >> 16))
	buf.WriteByte(byte(v >> 24))
}

func writeUint16LE(buf *bytes.Buffer, v uint16) {
	buf.WriteByte(byte(v))
	buf.WriteByte(byte(v >> 8))
}
