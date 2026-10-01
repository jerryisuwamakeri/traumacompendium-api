package moodservice

import (
	"context"
	"testing"
	"time"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func TestSummarize_ComputesAverageAndTagFrequency(t *testing.T) {
	svc := New(fakes.NewMoodRepo())
	ctx := context.Background()

	fixed := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixed }

	scores := []int{2, 4, 4}
	tagSets := [][]string{{"work stress"}, {"work stress", "poor sleep"}, {"poor sleep"}}
	for i, score := range scores {
		if _, err := svc.Create(ctx, "user-1", CreateInput{Score: score, Tags: tagSets[i]}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	summary, err := svc.Summarize(ctx, "user-1", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}

	if summary.EntryCount != 3 {
		t.Errorf("EntryCount = %d, want 3", summary.EntryCount)
	}
	wantAvg := (2.0 + 4.0 + 4.0) / 3.0
	if summary.AverageScore != roundToTwoDecimals(wantAvg) {
		t.Errorf("AverageScore = %v, want %v", summary.AverageScore, roundToTwoDecimals(wantAvg))
	}

	if len(summary.TagFrequency) != 2 {
		t.Fatalf("TagFrequency = %+v, want 2 distinct tags", summary.TagFrequency)
	}
	// Both tags appear twice; "poor sleep" sorts before "work stress"
	// alphabetically as the tie-break.
	if summary.TagFrequency[0].Tag != "poor sleep" || summary.TagFrequency[0].Count != 2 {
		t.Errorf("TagFrequency[0] = %+v, want {poor sleep 2}", summary.TagFrequency[0])
	}
	if summary.TagFrequency[1].Tag != "work stress" || summary.TagFrequency[1].Count != 2 {
		t.Errorf("TagFrequency[1] = %+v, want {work stress 2}", summary.TagFrequency[1])
	}
}

func TestSummarize_DefaultsToTrailing30Days(t *testing.T) {
	svc := New(fakes.NewMoodRepo())
	ctx := context.Background()

	fixed := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixed }

	summary, err := svc.Summarize(ctx, "user-1", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}

	wantFrom := fixed.Add(-defaultSummaryWindow)
	if !summary.From.Equal(wantFrom) {
		t.Errorf("From = %v, want %v", summary.From, wantFrom)
	}
	if !summary.To.Equal(fixed) {
		t.Errorf("To = %v, want %v", summary.To, fixed)
	}
}

func TestSummarize_EmptyRange_ReturnsZeroedSummary(t *testing.T) {
	svc := New(fakes.NewMoodRepo())

	summary, err := svc.Summarize(context.Background(), "user-1", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}
	if summary.EntryCount != 0 || summary.AverageScore != 0 {
		t.Errorf("summary = %+v, want zeroed", summary)
	}
	if summary.TagFrequency == nil {
		t.Error("TagFrequency should be an empty slice, not nil, for a clean JSON [] response")
	}
}

func TestSummarize_FromAfterTo_ReturnsValidationError(t *testing.T) {
	svc := New(fakes.NewMoodRepo())

	to := time.Now()
	from := to.Add(1 * time.Hour)

	_, err := svc.Summarize(context.Background(), "user-1", from, to)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeValidation {
		t.Errorf("Summarize() error = %v, want validation_error", err)
	}
}

func TestSummarize_OnlyIncludesCallersEntries(t *testing.T) {
	svc := New(fakes.NewMoodRepo())
	ctx := context.Background()

	if _, err := svc.Create(ctx, "user-1", CreateInput{Score: 5}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Create(ctx, "user-2", CreateInput{Score: 1}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	summary, err := svc.Summarize(ctx, "user-1", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}
	if summary.EntryCount != 1 || summary.AverageScore != 5 {
		t.Errorf("summary = %+v, want just user-1's single score-5 entry", summary)
	}
}
