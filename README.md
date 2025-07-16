# FirebaseManagementService

금융권 수준의 보안을 위한 서버 기반 FCM 토큰 발급 및 푸시 알림 관리 시스템

## 🔐 보안 특징

- **서버 기반 토큰 발급**: 클라이언트가 아닌 서버에서 직접 FCM 토큰 발급
- **장치 정보 검증**: 상세한 장치 정보를 통한 토큰 발급 및 관리
- **토큰 만료 관리**: 자동 만료 및 갱신 시스템
- **Redis 기반 저장**: 빠른 토큰 조회 및 관리

## 🚀 API 엔드포인트

### 1. 서버 기반 FCM 토큰 발급
```http
POST /fcm/generate
Content-Type: application/json

{
  "device_info": {
    "user_id": "user123",
    "platform": "android",
    "device_model": "Samsung Galaxy S23",
    "os_version": "Android 13",
    "app_version": "1.0.0",
    "device_id": "unique_device_identifier",
    "installation_id": "unique_installation_id"
  }
}
```

**응답:**
```json
{
  "token": "fcm_generated_token_here",
  "device_info": {
    "user_id": "user123",
    "platform": "android",
    "device_model": "Samsung Galaxy S23",
    "os_version": "Android 13",
    "app_version": "1.0.0",
    "device_id": "unique_device_identifier",
    "installation_id": "unique_installation_id"
  },
  "generated_at": "2024-01-01T00:00:00Z",
  "expires_at": "2025-01-01T00:00:00Z"
}
```

### 2. 푸시 알림 전송
```http
POST /fcm/send
Content-Type: application/json

{
  "user_id": "user123",
  "platform": "android",
  "title": "보안 알림",
  "body": "새로운 로그인이 감지되었습니다."
}
```

### 3. Firebase Custom Token 발급
```http
POST /auth/token
Content-Type: application/json

{
  "user_id": "user123"
}
```

## 🔧 환경 설정

```bash
# .env 파일
FIREBASE_CREDENTIALS_PATH=/path/to/firebase-credentials.json
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_redis_password
```

## 📱 클라이언트 사용 예시

### Android (Kotlin)
```kotlin
// 장치 정보 수집
val deviceInfo = DeviceInfo(
    userId = "user123",
    platform = "android",
    deviceModel = Build.MODEL,
    osVersion = Build.VERSION.RELEASE,
    appVersion = BuildConfig.VERSION_NAME,
    deviceId = Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID),
    installationId = getInstallationId() // 앱 설치별 고유 ID
)

// 서버에서 토큰 발급 요청
val response = apiService.generateFCMToken(deviceInfo)
val fcmToken = response.token

// Firebase에 토큰 등록
FirebaseMessaging.getInstance().token.addOnCompleteListener { task ->
    if (task.isSuccessful) {
        // 서버에서 발급받은 토큰을 Firebase에 등록
        FirebaseMessaging.getInstance().token = fcmToken
    }
}
```

### iOS (Swift)
```swift
// 장치 정보 수집
let deviceInfo = DeviceInfo(
    userId: "user123",
    platform: "ios",
    deviceModel: UIDevice.current.model,
    osVersion: UIDevice.current.systemVersion,
    appVersion: Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "",
    deviceId: UIDevice.current.identifierForVendor?.uuidString ?? "",
    installationId: getInstallationId() // 앱 설치별 고유 ID
)

// 서버에서 토큰 발급 요청
apiService.generateFCMToken(deviceInfo: deviceInfo) { result in
    switch result {
    case .success(let response):
        // 서버에서 발급받은 토큰을 Firebase에 등록
        Messaging.messaging().apnsToken = response.token.data(using: .utf8)
    case .failure(let error):
        print("Token generation failed: \(error)")
    }
}
```

## 🔒 보안 고려사항

1. **장치 정보 검증**: 모든 필수 장치 정보가 제공되어야 함
2. **토큰 만료**: 1년 후 자동 만료
3. **Redis 보안**: Redis 접근 제한 및 암호화
4. **Firebase 인증**: 서비스 계정 키 보안 관리
5. **HTTPS 통신**: 모든 API 통신은 HTTPS 필수

## 🛠️ 설치 및 실행

```bash
# 의존성 설치
go mod tidy

# 환경 변수 설정
cp .env.example .env
# .env 파일 편집

# 서버 실행
go run cmd/main.go
```

## 📊 모니터링

- 토큰 발급 로그
- 푸시 전송 성공/실패 통계
- 장치별 토큰 관리 현황
- Redis 저장소 사용량 모니터링
