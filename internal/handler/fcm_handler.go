package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Jseongwon/FirebaseManagementService.git/internal/fcm"
	"github.com/Jseongwon/FirebaseManagementService.git/internal/redis"
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

// 서버에서 FCM 토큰을 발급하는 새로운 핸들러
func GenerateFCMToken(w http.ResponseWriter, r *http.Request) {
	log.Println("📥 [Start] GenerateFCMToken handler called")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Failed to read request body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	log.Println("✅ Request body read")

	var req fcm.TokenGenerationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("❌ JSON unmarshal failed: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Printf("✅ Request parsed: %+v", req.DeviceInfo)

	// 장치 정보 검증
	if err := validateDeviceInfo(req.DeviceInfo); err != nil {
		log.Printf("❌ Device info validation failed: %v", err)
		http.Error(w, "Invalid device info", http.StatusBadRequest)
		return
	}

	// FCM 토큰 발급
	tokenResponse, err := fcm.GenerateFCMToken(req.DeviceInfo)
	if err != nil {
		log.Printf("❌ FCM token generation failed: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Redis에 토큰 저장 (장치 정보 포함)
	tokenInfo := redis.TokenInfo{
		Token:          tokenResponse.Token,
		UserID:         req.DeviceInfo.UserID,
		Platform:       req.DeviceInfo.Platform,
		DeviceModel:    req.DeviceInfo.DeviceModel,
		OSVersion:      req.DeviceInfo.OSVersion,
		AppVersion:     req.DeviceInfo.AppVersion,
		DeviceID:       req.DeviceInfo.DeviceID,
		InstallationID: req.DeviceInfo.InstallationID,
		GeneratedAt:    time.Now(),
		ExpiresAt:      time.Now().AddDate(1, 0, 0), // 1년 후 만료
	}

	err = redisClient.SetFCMTokenWithDeviceInfo(r.Context(), tokenInfo)
	if err != nil {
		log.Printf("❌ Failed to store token in Redis: %v", err)
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	// 응답 전송
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse)

	log.Printf("✅ FCM token generated and stored. UserID=%s, Platform=%s", req.DeviceInfo.UserID, req.DeviceInfo.Platform)
}

// 장치 정보 검증 함수
func validateDeviceInfo(deviceInfo fcm.DeviceInfo) error {
	if deviceInfo.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if deviceInfo.Platform == "" {
		return fmt.Errorf("platform is required")
	}
	if deviceInfo.DeviceModel == "" {
		return fmt.Errorf("device_model is required")
	}
	if deviceInfo.OSVersion == "" {
		return fmt.Errorf("os_version is required")
	}
	if deviceInfo.DeviceID == "" {
		return fmt.Errorf("device_id is required")
	}
	if deviceInfo.InstallationID == "" {
		return fmt.Errorf("installation_id is required")
	}

	// 플랫폼 검증
	validPlatforms := map[string]bool{
		"android": true,
		"ios":     true,
		"web":     true,
	}
	if !validPlatforms[deviceInfo.Platform] {
		return fmt.Errorf("invalid platform: %s", deviceInfo.Platform)
	}

	return nil
}
