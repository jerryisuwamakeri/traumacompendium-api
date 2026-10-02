package fakes

import (
	"context"
	"sync"

	"github.com/traumacompendium/backend/internal/notification/tts"
)

// TTSSynthesizer is an in-memory tts.Synthesizer that records every
// request it's asked to synthesize and returns deterministic fake audio,
// so tests can assert both on the synthesized output and on how many
// times synthesis was actually invoked (e.g. to prove a cache hit skips
// it).
type TTSSynthesizer struct {
	mu    sync.Mutex
	Calls []tts.Request
}

func (f *TTSSynthesizer) Synthesize(_ context.Context, req tts.Request) (tts.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, req)
	return tts.Result{Audio: []byte("fake-audio:" + req.Locale), ContentType: "audio/mpeg"}, nil
}

func (f *TTSSynthesizer) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Calls)
}
