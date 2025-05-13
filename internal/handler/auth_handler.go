package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Jseongwon/FirebaseManagementService.git/internal/firebase"
)

type TokenRequest struct {
	UserID string `json:"user_id"`
}

func IssueCustomToken(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := firebase.CreateCustomToken(req.UserID)
	if err != nil {
		http.Error(w, "Token creation error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"custom_token": token})
}
