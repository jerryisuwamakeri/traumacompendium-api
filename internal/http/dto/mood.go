package dto

import (
	"time"

	"github.com/traumacompendium/backend/internal/domain/mood"
	"github.com/traumacompendium/backend/internal/service/moodservice"
)

type CreateMoodEntryRequest struct {
	Score int      `json:"score" binding:"required,min=1,max=5"`
	Tags  []string `json:"tags" binding:"max=20,dive,max=50"`
	Note  string   `json:"note" binding:"max=2000"`
}

type MoodEntryResponse struct {
	ID        string    `json:"id"`
	Score     int       `json:"score"`
	Tags      []string  `json:"tags"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func NewMoodEntryResponse(e *mood.Entry) MoodEntryResponse {
	return MoodEntryResponse{
		ID:        e.ID,
		Score:     e.Score,
		Tags:      e.Tags,
		Note:      e.Note,
		CreatedAt: e.CreatedAt,
	}
}

func NewMoodEntryResponses(entries []*mood.Entry) []MoodEntryResponse {
	out := make([]MoodEntryResponse, len(entries))
	for i, e := range entries {
		out[i] = NewMoodEntryResponse(e)
	}
	return out
}

type TagCountResponse struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type MoodSummaryResponse struct {
	From         time.Time          `json:"from"`
	To           time.Time          `json:"to"`
	EntryCount   int                `json:"entry_count"`
	AverageScore float64            `json:"average_score"`
	TagFrequency []TagCountResponse `json:"tag_frequency"`
}

func NewMoodSummaryResponse(s moodservice.Summary) MoodSummaryResponse {
	tags := make([]TagCountResponse, len(s.TagFrequency))
	for i, tc := range s.TagFrequency {
		tags[i] = TagCountResponse{Tag: tc.Tag, Count: tc.Count}
	}
	return MoodSummaryResponse{
		From:         s.From,
		To:           s.To,
		EntryCount:   s.EntryCount,
		AverageScore: s.AverageScore,
		TagFrequency: tags,
	}
}
