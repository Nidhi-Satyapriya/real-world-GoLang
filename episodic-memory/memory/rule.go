package memory

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// BehaviorRule is a reusable correction extracted from human feedback.
// It captures the pattern that triggered the correction, the corrective
// action, and contextual metadata used for retrieval.
type BehaviorRule struct {
	ID        string            `json:"id"`
	Pattern   string            `json:"pattern"`
	Action    string            `json:"action"`
	Domain    string            `json:"domain,omitempty"`
	Task      string            `json:"task,omitempty"`
	Source    string            `json:"source,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Embedding []float64         `json:"embedding,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// SearchResult pairs a retrieved rule with its similarity score (0..1).
type SearchResult struct {
	Rule  BehaviorRule `json:"rule"`
	Score float64      `json:"score"`
}

func NewRuleID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}
