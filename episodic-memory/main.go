package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"episodic-memory/extractor"
	"episodic-memory/memory"
)

var (
	store     memory.MemoryStore
	ext       *extractor.Extractor
	startTime time.Time
)

func main() {
	embedder := memory.NewTFIDFEmbedder()
	store = memory.NewVectorStore(embedder, "memory.json")
	ext = extractor.New()
	startTime = time.Now()

	mux := http.NewServeMux()

	mux.HandleFunc("/corrections", handleCorrections)
	mux.HandleFunc("/recall", handleRecall)
	mux.HandleFunc("/rules", handleRules)
	mux.HandleFunc("/rules/", handleRuleByID)
	mux.HandleFunc("/reindex", handleReindex)
	mux.HandleFunc("/health", handleHealth)

	addr := ":8090"
	log.Printf("[EPISODIC-MEMORY] Listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[EPISODIC-MEMORY] Server failed: %v", err)
	}
}

// POST /corrections — submit a human correction, extract a rule, store it
func handleCorrections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var input extractor.CorrectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid json: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Raw) == "" {
		http.Error(w, `{"error":"raw correction text is required"}`, http.StatusBadRequest)
		return
	}

	rule := ext.Extract(input)
	if err := store.Store(rule); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"storage failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	log.Printf("[CORRECTION] stored rule %s domain=%s task=%s", rule.ID, rule.Domain, rule.Task)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "stored",
		"rule":   rule,
	})
}

// GET /recall?query=...&top_k=5&domain=...&task=...
// The core retrieval endpoint — an agent calls this before starting work.
func handleRecall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, `{"error":"query parameter is required"}`, http.StatusBadRequest)
		return
	}

	topK := 5
	if k := r.URL.Query().Get("top_k"); k != "" {
		fmt.Sscanf(k, "%d", &topK)
	}

	filters := make(map[string]string)
	if d := r.URL.Query().Get("domain"); d != "" {
		filters["domain"] = d
	}
	if t := r.URL.Query().Get("task"); t != "" {
		filters["task"] = t
	}

	results := store.Query(query, topK, filters)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	})
}

// GET /rules — list all stored rules
func handleRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	rules := store.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(rules),
		"rules": rules,
	})
}

// GET/DELETE /rules/{id}
func handleRuleByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/rules/")
	if id == "" {
		http.Error(w, `{"error":"rule id required"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rule, ok := store.Get(id)
		if !ok {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rule)

	case http.MethodDelete:
		if !store.Delete(id) {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": id})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// POST /reindex — re-embed all rules after vocabulary growth
func handleReindex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	store.Reindex()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reindexed"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"uptime": time.Since(startTime).String(),
		"rules":  len(store.List()),
	})
}
