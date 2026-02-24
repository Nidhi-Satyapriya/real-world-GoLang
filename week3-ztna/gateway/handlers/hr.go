package handlers

import (
	"encoding/json"
	"net/http"

	"week3-ztna-gateway/middleware"
)

func HRResource(w http.ResponseWriter, r *http.Request) {
	if err := r.Context().Err(); err != nil {
		http.Error(w, `{"error":"request timed out"}`, http.StatusGatewayTimeout)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	role, _ := r.Context().Value(middleware.RoleKey).(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from HR — Confidential Zone",
		"user":    userID,
		"role":    role,
	})
}
