package moodservice

import (
	"context"
	"sort"
	"time"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

// Summary aggregates a user's mood entries over a date range — enough
// for a simple trend view (brief: recovery tools should surface
// "progress") without needing a dedicated analytics pipeline.
type Summary struct {
	From         time.Time
	To           time.Time
	EntryCount   int
	AverageScore float64
	// TagFrequency counts how often each (normalized) trigger tag
	// appears, so the most common triggers surface first.
	TagFrequency []TagCount
}

type TagCount struct {
	Tag   string
	Count int
}

const defaultSummaryWindow = 30 * 24 * time.Hour

// Summarize computes a Summary for userID over [from, to]. A zero from/to
// defaults the window to the trailing 30 days ending now, matching most
// "recent trend" product expectations without the caller needing to
// compute dates itself.
func (s *Service) Summarize(ctx context.Context, userID string, from, to time.Time) (Summary, error) {
	if to.IsZero() {
		to = s.now()
	}
	if from.IsZero() {
		from = to.Add(-defaultSummaryWindow)
	}
	if from.After(to) {
		return Summary{}, apperror.NewValidation("from must not be after to")
	}

	entries, err := s.entries.ListInRange(ctx, userID, from, to)
	if err != nil {
		return Summary{}, apperror.NewInternal("could not summarize mood entries").WithCause(err)
	}

	summary := Summary{From: from, To: to, EntryCount: len(entries)}
	if len(entries) == 0 {
		summary.TagFrequency = []TagCount{}
		return summary, nil
	}

	tagCounts := make(map[string]int)
	var scoreSum int
	for _, e := range entries {
		scoreSum += e.Score
		for _, tag := range e.Tags {
			tagCounts[tag]++
		}
	}
	summary.AverageScore = roundToTwoDecimals(float64(scoreSum) / float64(len(entries)))
	summary.TagFrequency = sortedTagCounts(tagCounts)

	return summary, nil
}

func sortedTagCounts(counts map[string]int) []TagCount {
	out := make([]TagCount, 0, len(counts))
	for tag, count := range counts {
		out = append(out, TagCount{Tag: tag, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Tag < out[j].Tag // stable, deterministic tie-break
	})
	return out
}

func roundToTwoDecimals(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
