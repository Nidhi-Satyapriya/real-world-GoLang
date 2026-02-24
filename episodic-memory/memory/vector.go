package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

// MemoryStore is the interface any backing store (in-memory, Pinecone, Milvus)
// must implement. The rest of the system programs against this contract.
type MemoryStore interface {
	Store(rule BehaviorRule) error
	Query(queryText string, topK int, filters map[string]string) []SearchResult
	Get(id string) (BehaviorRule, bool)
	Delete(id string) bool
	List() []BehaviorRule
	Reindex()
}

// VectorStore is an in-memory implementation of MemoryStore backed by
// cosine similarity over TF-IDF embeddings. Swap for Pinecone/Milvus
// by implementing the MemoryStore interface.
type VectorStore struct {
	mu       sync.RWMutex
	rules    map[string]BehaviorRule
	embedder *TFIDFEmbedder
	persist  string // file path for JSON persistence; empty = no persistence
}

func NewVectorStore(embedder *TFIDFEmbedder, persistPath string) *VectorStore {
	vs := &VectorStore{
		rules:    make(map[string]BehaviorRule),
		embedder: embedder,
		persist:  persistPath,
	}
	if persistPath != "" {
		vs.loadFromDisk()
	}
	return vs
}

func (vs *VectorStore) Store(rule BehaviorRule) error {
	combined := rule.Pattern + " " + rule.Action + " " + rule.Domain + " " + rule.Task
	vs.embedder.Train(combined)
	rule.Embedding = vs.embedder.Embed(combined)

	if rule.ID == "" {
		rule.ID = NewRuleID()
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now().UTC()
	}

	vs.mu.Lock()
	vs.rules[rule.ID] = rule
	vs.mu.Unlock()

	vs.saveToDisk()
	return nil
}

// Query returns the topK most similar rules. Optional filters narrow by
// exact-match on domain, task, or any metadata key.
func (vs *VectorStore) Query(queryText string, topK int, filters map[string]string) []SearchResult {
	queryVec := vs.embedder.Embed(queryText)

	vs.mu.RLock()
	defer vs.mu.RUnlock()

	var results []SearchResult
	for _, rule := range vs.rules {
		if !matchesFilters(rule, filters) {
			continue
		}
		score := CosineSimilarity(queryVec, rule.Embedding)
		results = append(results, SearchResult{Rule: rule, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results
}

func (vs *VectorStore) Get(id string) (BehaviorRule, bool) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	r, ok := vs.rules[id]
	return r, ok
}

func (vs *VectorStore) Delete(id string) bool {
	vs.mu.Lock()
	_, existed := vs.rules[id]
	delete(vs.rules, id)
	vs.mu.Unlock()
	if existed {
		vs.saveToDisk()
	}
	return existed
}

func (vs *VectorStore) List() []BehaviorRule {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	out := make([]BehaviorRule, 0, len(vs.rules))
	for _, r := range vs.rules {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

// Reindex re-embeds every stored rule. Call after the vocabulary has grown
// significantly (e.g., after a batch of new rules).
func (vs *VectorStore) Reindex() {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	for id, rule := range vs.rules {
		combined := rule.Pattern + " " + rule.Action + " " + rule.Domain + " " + rule.Task
		rule.Embedding = vs.embedder.Embed(combined)
		vs.rules[id] = rule
	}
	vs.saveToDiskLocked()
}

// ---------------------------------------------------------------------------
// Persistence (JSON file — simple, portable, no external deps)
// ---------------------------------------------------------------------------

type snapshot struct {
	Rules    []BehaviorRule `json:"rules"`
	Vocab    map[string]int `json:"vocab"`
	DocFreq  map[string]int `json:"doc_freq"`
	DocCount int            `json:"doc_count"`
}

func (vs *VectorStore) saveToDisk() {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	vs.saveToDiskLocked()
}

func (vs *VectorStore) saveToDiskLocked() {
	if vs.persist == "" {
		return
	}
	snap := snapshot{
		Rules:    make([]BehaviorRule, 0, len(vs.rules)),
		DocCount: vs.embedder.docCount,
	}
	for _, r := range vs.rules {
		snap.Rules = append(snap.Rules, r)
	}
	vs.embedder.mu.RLock()
	snap.Vocab = vs.embedder.vocab
	snap.DocFreq = vs.embedder.docFreq
	vs.embedder.mu.RUnlock()

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[memory] persist error: %v\n", err)
		return
	}
	if err := os.WriteFile(vs.persist, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "[memory] persist write error: %v\n", err)
	}
}

func (vs *VectorStore) loadFromDisk() {
	data, err := os.ReadFile(vs.persist)
	if err != nil {
		return // file doesn't exist yet
	}
	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		fmt.Fprintf(os.Stderr, "[memory] corrupt snapshot: %v\n", err)
		return
	}

	vs.embedder.mu.Lock()
	vs.embedder.vocab = snap.Vocab
	vs.embedder.docFreq = snap.DocFreq
	vs.embedder.docCount = snap.DocCount
	vs.embedder.mu.Unlock()

	for _, r := range snap.Rules {
		vs.rules[r.ID] = r
	}
	fmt.Printf("[memory] loaded %d rules from %s\n", len(vs.rules), vs.persist)
}

// ---------------------------------------------------------------------------
// Filter helpers
// ---------------------------------------------------------------------------

func matchesFilters(rule BehaviorRule, filters map[string]string) bool {
	for key, val := range filters {
		switch key {
		case "domain":
			if rule.Domain != val {
				return false
			}
		case "task":
			if rule.Task != val {
				return false
			}
		case "source":
			if rule.Source != val {
				return false
			}
		default:
			if rule.Metadata == nil || rule.Metadata[key] != val {
				return false
			}
		}
	}
	return true
}
