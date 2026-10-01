package dto

import (
	"time"

	"github.com/traumacompendium/backend/internal/domain/journal"
)

type CreateJournalEntryRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,max=20000"`
}

type UpdateJournalEntryRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,max=20000"`
}

type JournalEntryResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewJournalEntryResponse(e *journal.Entry) JournalEntryResponse {
	return JournalEntryResponse{
		ID:        e.ID,
		Title:     e.Title,
		Content:   e.Content,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func NewJournalEntryResponses(entries []*journal.Entry) []JournalEntryResponse {
	out := make([]JournalEntryResponse, len(entries))
	for i, e := range entries {
		out[i] = NewJournalEntryResponse(e)
	}
	return out
}
