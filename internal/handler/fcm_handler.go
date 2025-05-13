package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Jseongwon/FirebaseManagementService.git/internal/fcm"
)

type RegisterFCMRequest struct {
	UserID   string `json:"user_id"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

func RegisterFCMToken(w http.ResponseWriter, r *http.Request) {
	var req RegisterFCMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := redisClient.SetFCMToken(r.Context(), req.UserID, req.Platform, req.Token)
	if err != nil {
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type SendNotificationRequest struct {
	UserID   string `json:"user_id"`
	Platform string `json:"platform"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

// ✔️ 정확한 이름과 핸들러 시그니처 유지
func SendNotification(w http.ResponseWriter, r *http.Request) {
	log.Println("📥 [Start] SendNotification handler called")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Failed to read request body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	log.Println("✅ Request body read")

	var req SendNotificationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("❌ JSON unmarshal failed: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Printf("✅ Request parsed: %+v", req)

	token, err := redisClient.GetFCMToken(r.Context(), req.UserID, req.Platform)
	if err != nil {
		log.Printf("❌ Token fetch failed: %v", err)
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}
	log.Printf("✅ Token fetched: %s", token)

	err = fcm.SendNotification(token, req.Title, req.Body)
	if err != nil {
		log.Printf("❌ FCM send failed: %v", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Notification sent. To=%s, Title=%s, Body=%s", token, req.Title, req.Body)
	w.WriteHeader(http.StatusOK)
}
