// Package assistantservice implements the AI assistant's reply
// pipeline: a Premium-gated check, a constrained system prompt limiting
// replies to educational trauma content (never clinical advice — the
// brief's own named risk-mitigation), a bounded conversation-history
// context window, and a hard backstop that skips the model entirely for
// messages containing crisis language, returning a fixed
// crisis-resources response instead. That backstop exists because
// prompt-following alone is not an acceptable safety margin for
// self-harm/suicide content — see containsCrisisLanguage.
package assistantservice

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/assistant"
	"github.com/traumacompendium/backend/internal/notification/ai"
)

// contextWindowSize bounds how many of a conversation's most recent
// messages are sent to the model — a long-running conversation's prompt
// never grows without bound.
const contextWindowSize = 20

const systemPrompt = `You are the TraumaCompendium AI assistant. You provide general, educational information about trauma, recovery, and mental wellbeing — the kind of material covered in this app's learning content. You are not a therapist, counselor, or medical professional: never provide a diagnosis, a clinical treatment plan, or crisis intervention, and never claim clinical authority. If a user describes a medical or mental health emergency, tell them clearly to contact local emergency services or a crisis line immediately rather than trying to help them through it yourself. Keep replies concise, supportive, and educational, and make clear you are an AI assistant, not a substitute for professional care.`

// crisisResponse is returned verbatim, without involving the model at
// all, whenever containsCrisisLanguage matches — deliberately generic
// about jurisdiction, since this service doesn't know the user's
// location.
const crisisResponse = "I'm really concerned about what you've shared, and I want you to be safe. I'm not able to provide crisis support myself, but please reach out right now to a crisis line or your local emergency services if you're in danger — don't wait. If you're not in immediate danger but are struggling, please also consider reaching out to someone you trust or a mental health professional. You deserve support from people who can actually help you in real time."

// crisisKeywords is a blunt, deliberately narrow safety net — not a
// comprehensive crisis-detection system. Matching is a simple
// case-insensitive substring check against common self-harm/suicide
// phrasing.
var crisisKeywords = []string{
	"kill myself",
	"suicide",
	"want to die",
	"end my life",
	"ending my life",
	"hurt myself",
	"harm myself",
	"self-harm",
	"no reason to live",
	"better off dead",
}

func containsCrisisLanguage(text string) bool {
	lower := strings.ToLower(text)
	for _, kw := range crisisKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// PremiumChecker is the one thing this service needs from billing — a
// narrow interface (billingservice.Service satisfies it without an
// adapter) so this package doesn't depend on billingservice's full
// surface just to gate one feature.
type PremiumChecker interface {
	IsPremium(ctx context.Context, userID string) (bool, error)
}

type Service struct {
	repo    assistant.Repository
	ai      ai.Client
	premium PremiumChecker
	log     zerolog.Logger
	now     func() time.Time
}

func New(repo assistant.Repository, aiClient ai.Client, premium PremiumChecker, log zerolog.Logger) *Service {
	return &Service{repo: repo, ai: aiClient, premium: premium, log: log, now: time.Now}
}
