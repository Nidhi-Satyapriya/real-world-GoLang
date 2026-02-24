package memory

import (
	"math"
	"strings"
	"sync"
	"unicode"
)

// Embedder turns text into a dense vector representation.
// Swap TFIDFEmbedder for an OpenAI/Cohere adapter in production.
type Embedder interface {
	Embed(text string) []float64
	Dimensions() int
}

// ---------------------------------------------------------------------------
// TF-IDF Embedder — fully local, no external API calls
// ---------------------------------------------------------------------------
// Builds a growing vocabulary from every text it sees. Each Embed call returns
// a TF-IDF-weighted vector in that vocabulary space. Good enough for semantic
// recall over hundreds-to-low-thousands of rules without any network calls.

type TFIDFEmbedder struct {
	mu       sync.RWMutex
	vocab    map[string]int // token -> column index
	docFreq  map[string]int // token -> number of documents containing it
	docCount int
}

func NewTFIDFEmbedder() *TFIDFEmbedder {
	return &TFIDFEmbedder{
		vocab:   make(map[string]int),
		docFreq: make(map[string]int),
	}
}

func (e *TFIDFEmbedder) Dimensions() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.vocab)
}

// Train updates vocabulary and document-frequency counts. Call this when a new
// rule is stored so that future Embed calls benefit from the updated IDF.
func (e *TFIDFEmbedder) Train(text string) {
	tokens := tokenize(text)
	seen := make(map[string]bool)

	e.mu.Lock()
	defer e.mu.Unlock()

	e.docCount++
	for _, t := range tokens {
		if _, exists := e.vocab[t]; !exists {
			e.vocab[t] = len(e.vocab)
		}
		if !seen[t] {
			e.docFreq[t]++
			seen[t] = true
		}
	}
}

// Embed returns a TF-IDF vector for the given text. The vector length equals
// the current vocabulary size; callers must re-embed stored rules when the
// vocabulary grows (VectorStore.Reindex handles this).
func (e *TFIDFEmbedder) Embed(text string) []float64 {
	tokens := tokenize(text)
	tf := make(map[string]int)
	for _, t := range tokens {
		tf[t]++
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	vec := make([]float64, len(e.vocab))
	for token, count := range tf {
		idx, ok := e.vocab[token]
		if !ok {
			continue
		}
		termFreq := float64(count) / float64(len(tokens))
		idf := math.Log(float64(e.docCount+1) / float64(e.docFreq[token]+1))
		vec[idx] = termFreq * idf
	}

	normalize(vec)
	return vec
}

// ---------------------------------------------------------------------------
// Similarity
// ---------------------------------------------------------------------------

func CosineSimilarity(a, b []float64) float64 {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	var dot, normA, normB float64
	for i := 0; i < minLen; i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	for i := minLen; i < len(a); i++ {
		normA += a[i] * a[i]
	}
	for i := minLen; i < len(b); i++ {
		normB += b[i] * b[i]
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// ---------------------------------------------------------------------------
// Text helpers
// ---------------------------------------------------------------------------

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	word := strings.Builder{}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			word.WriteRune(r)
		} else if word.Len() > 0 {
			t := word.String()
			if !isStopWord(t) {
				tokens = append(tokens, t)
			}
			word.Reset()
		}
	}
	if word.Len() > 0 {
		t := word.String()
		if !isStopWord(t) {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

func normalize(vec []float64) {
	var sum float64
	for _, v := range vec {
		sum += v * v
	}
	if norm := math.Sqrt(sum); norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}
}

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "it": true,
	"in": true, "on": true, "at": true, "to": true, "for": true,
	"of": true, "and": true, "or": true, "but": true, "not": true,
	"with": true, "this": true, "that": true, "from": true, "by": true,
	"be": true, "as": true, "are": true, "was": true, "were": true,
	"been": true, "has": true, "have": true, "had": true, "do": true,
	"does": true, "did": true, "will": true, "would": true, "should": true,
	"can": true, "could": true, "if": true, "then": true, "so": true,
}

func isStopWord(w string) bool {
	return stopWords[w]
}
