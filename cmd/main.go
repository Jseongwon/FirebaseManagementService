package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Jseongwon/FirebaseManagementService.git/internal/firebase"
	"github.com/Jseongwon/FirebaseManagementService.git/internal/handler"
	"github.com/Jseongwon/FirebaseManagementService.git/internal/redis"
	"github.com/joho/godotenv"
)

func main() {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Using system environment variables.")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	log.Println("redisAddr: ", redisAddr)
	log.Println("redisPassword: ", redisPassword)
	log.Println("redisDB: ", redisDB)

	redisClient := redis.NewRedisClient(redisAddr, redisPassword, redisDB)

	firebasePath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	firebase.InitFirebase(firebasePath)

	handler.InitHandler(redisClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/token", handler.IssueCustomToken)
	mux.HandleFunc("/fcm/generate", handler.GenerateFCMToken) // 새로운 서버 기반 토큰 발급
	mux.HandleFunc("/fcm/register", handler.RegisterFCMToken) // 기존 클라이언트 토큰 등록
	mux.HandleFunc("/fcm/send", handler.SendNotification)

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", mux)
}
