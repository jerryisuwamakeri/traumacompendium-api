package learningservice

import (
	"context"
	"testing"
)

func TestGetLessonAudio_CacheMiss_SynthesizesAndCaches(t *testing.T) {
	svc, content, _, audio, synth := newTestServiceWithAudio()
	_, lesson1ID, _, _, _, _, _ := seedTwoLessonLevel(content)

	result, err := svc.GetLessonAudio(context.Background(), lesson1ID, "fr")
	if err != nil {
		t.Fatalf("GetLessonAudio() error = %v", err)
	}
	if len(result.Data) == 0 {
		t.Error("expected non-empty audio data")
	}
	if synth.CallCount() != 1 {
		t.Errorf("synth.CallCount() = %d, want 1 (cache miss should synthesize once)", synth.CallCount())
	}
	if audio.Count() != 1 {
		t.Errorf("audio.Count() = %d, want 1 (result should be cached)", audio.Count())
	}
}

func TestGetLessonAudio_CacheHit_DoesNotResynthesize(t *testing.T) {
	svc, content, _, _, synth := newTestServiceWithAudio()
	_, lesson1ID, _, _, _, _, _ := seedTwoLessonLevel(content)
	ctx := context.Background()

	first, err := svc.GetLessonAudio(ctx, lesson1ID, "fr")
	if err != nil {
		t.Fatalf("first GetLessonAudio() error = %v", err)
	}

	second, err := svc.GetLessonAudio(ctx, lesson1ID, "fr")
	if err != nil {
		t.Fatalf("second GetLessonAudio() error = %v", err)
	}

	if synth.CallCount() != 1 {
		t.Errorf("synth.CallCount() = %d, want 1 (second request should be a cache hit)", synth.CallCount())
	}
	if string(first.Data) != string(second.Data) {
		t.Error("cached audio differs from originally synthesized audio")
	}
}

func TestGetLessonAudio_DifferentLocales_CachedSeparately(t *testing.T) {
	svc, content, _, audio, synth := newTestServiceWithAudio()
	_, lesson1ID, _, _, _, _, _ := seedTwoLessonLevel(content)
	ctx := context.Background()

	if _, err := svc.GetLessonAudio(ctx, lesson1ID, "fr"); err != nil {
		t.Fatalf("GetLessonAudio(fr) error = %v", err)
	}
	if _, err := svc.GetLessonAudio(ctx, lesson1ID, "sw"); err != nil {
		t.Fatalf("GetLessonAudio(sw) error = %v", err)
	}

	if synth.CallCount() != 2 {
		t.Errorf("synth.CallCount() = %d, want 2 (two distinct locales, two syntheses)", synth.CallCount())
	}
	if audio.Count() != 2 {
		t.Errorf("audio.Count() = %d, want 2 (cached per-locale)", audio.Count())
	}
}

func TestGetLessonAudio_UnknownLesson_ReturnsNotFound(t *testing.T) {
	svc, _, _, _, _ := newTestServiceWithAudio()

	_, err := svc.GetLessonAudio(context.Background(), "does-not-exist", "fr")
	if err == nil {
		t.Fatal("GetLessonAudio() for unknown lesson error = nil, want error")
	}
}
