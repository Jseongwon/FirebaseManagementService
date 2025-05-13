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
