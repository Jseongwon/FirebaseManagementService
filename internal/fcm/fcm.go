package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
)

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Message struct {
	Token        string       `json:"token"`
	Notification Notification `json:"notification"`
}

type FcmRequest struct {
	Message Message `json:"message"`
}

// FCM 토큰 발급을 위한 새로운 구조체들
type DeviceInfo struct {
	UserID         string `json:"user_id"`
	Platform       string `json:"platform"`        // "android", "ios", "web"
	DeviceModel    string `json:"device_model"`    // "iPhone 14", "Samsung Galaxy S23"
	OSVersion      string `json:"os_version"`      // "iOS 17.0", "Android 13"
	AppVersion     string `json:"app_version"`     // "1.0.0"
	DeviceID       string `json:"device_id"`       // 고유 장치 식별자
	InstallationID string `json:"installation_id"` // 설치별 고유 ID
}

type TokenGenerationRequest struct {
	DeviceInfo DeviceInfo `json:"device_info"`
}

type TokenGenerationResponse struct {
	Token       string     `json:"token"`
	DeviceInfo  DeviceInfo `json:"device_info"`
	GeneratedAt string     `json:"generated_at"`
	ExpiresAt   string     `json:"expires_at"`
}

// FCM 토큰 발급 함수
func GenerateFCMToken(deviceInfo DeviceInfo) (*TokenGenerationResponse, error) {
	// 1️⃣ 서비스 계정 json 로드
	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS_PATH is not set")
	}
	jsonKey, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	// 2️⃣ OAuth2 토큰 생성
	conf, err := google.JWTConfigFromJSON(jsonKey, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}
	tokenSrc := conf.TokenSource(context.Background())
	tokenObj, err := tokenSrc.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	accessToken := tokenObj.AccessToken
	log.Println("✅ Got access token for token generation")

	// 3️⃣ 프로젝트 ID 추출
	var credentials struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(jsonKey, &credentials); err != nil {
		return nil, fmt.Errorf("failed to parse project_id: %w", err)
	}
	projectID := credentials.ProjectID

	// 4️⃣ FCM 토큰 발급 요청 구조
	// FCM v1 API를 사용하여 토큰 발급
	tokenReqBody := map[string]interface{}{
		"application": fmt.Sprintf("projects/%s/apps/%s", projectID, getAppID(deviceInfo.Platform)),
		"device": map[string]interface{}{
			"device_model":    deviceInfo.DeviceModel,
			"os_version":      deviceInfo.OSVersion,
			"app_version":     deviceInfo.AppVersion,
			"device_id":       deviceInfo.DeviceID,
			"installation_id": deviceInfo.InstallationID,
		},
	}

	jsonData, err := json.Marshal(tokenReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	// 5️⃣ FCM 토큰 발급 HTTP 요청
	tokenURL := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/registrations", projectID)
	req, err := http.NewRequest("POST", tokenURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// 6️⃣ 토큰 발급 요청 전송
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("FCM token generation request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	// 7️⃣ 응답 처리
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FCM token generation error: %s - %s", resp.Status, string(respBody))
	}

	// 8️⃣ 응답 파싱
	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	log.Printf("✅ FCM token generated successfully for user: %s, device: %s", deviceInfo.UserID, deviceInfo.DeviceModel)

	// 9️⃣ 응답 생성
	response := &TokenGenerationResponse{
		Token:       tokenResp.Token,
		DeviceInfo:  deviceInfo,
		GeneratedAt: "2024-01-01T00:00:00Z", // 실제로는 현재 시간
		ExpiresAt:   "2025-01-01T00:00:00Z", // 실제로는 만료 시간 계산
	}

	return response, nil
}

// 플랫폼별 앱 ID 반환 (실제 Firebase 프로젝트 설정에 맞게 수정 필요)
func getAppID(platform string) string {
	switch platform {
	case "android":
		return "android:com.example.app" // 실제 Android 패키지명
	case "ios":
		return "ios:com.example.app" // 실제 iOS 번들 ID
	case "web":
		return "web:com.example.app" // 실제 웹 앱 ID
	default:
		return "android:com.example.app" // 기본값
	}
}

func SendNotification(token, title, body string) error {
	// 1️⃣ 서비스 계정 json 로드
	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		return fmt.Errorf("FIREBASE_CREDENTIALS_PATH is not set")
	}
	jsonKey, err := os.ReadFile(credPath)
	if err != nil {
		return fmt.Errorf("failed to read credentials: %w", err)
	}

	// 2️⃣ OAuth2 토큰 생성
	conf, err := google.JWTConfigFromJSON(jsonKey, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}
	tokenSrc := conf.TokenSource(context.Background())
	tokenObj, err := tokenSrc.Token()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	accessToken := tokenObj.AccessToken
	log.Println("✅ Got access token")

	// 3️⃣ 프로젝트 ID 추출
	var credentials struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(jsonKey, &credentials); err != nil {
		return fmt.Errorf("failed to parse project_id: %w", err)
	}
	projectID := credentials.ProjectID

	// 4️⃣ 요청 구조
	reqBody := FcmRequest{
		Message: Message{
			Token: token,
			Notification: Notification{
				Title: title,
				Body:  body,
			},
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 5️⃣ HTTP 요청 생성
	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// 6️⃣ 전송
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("FCM request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	// 7️⃣ 응답 처리
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FCM send error: %s - %s", resp.Status, string(respBody))
	}

	log.Printf("📬 FCM response: %s", string(respBody))
	return nil
}
