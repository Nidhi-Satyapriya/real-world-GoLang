package main

import (
	"log"
	"net/http"
	"time"

	"week3-ztna-gateway/handlers"
	"week3-ztna-gateway/middleware"
)

func main() {
	mux := http.NewServeMux()

	// Health check — no auth required
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Protected HR resource — full middleware chain
	hrHandler := middleware.Chain(
		http.HandlerFunc(handlers.HRResource),
		middleware.AccessLogger,
		middleware.RequestTimeout(5*time.Second),
		middleware.JWTAuth,
	)
	mux.Handle("/hr", hrHandler)

	addr := ":8080"
	log.Printf("[GATEWAY] Zero Trust Enforcer listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[GATEWAY] Server failed: %v", err)
	}
}
