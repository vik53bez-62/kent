# Kent — skeleton

Готовый скелет мессенджера Kent:

- Android (Kotlin/Compose/Hilt), package: com.kent.app
- Сервер (Go + Gin + Redis), OTP (Infobip)
- Docker Compose (Postgres, Redis, NATS, MinIO — можно расширять)
- CI (GitHub Actions), черновики документов для Google Play

Быстрый старт:

1) Сервер (dev):
   - cp server/.env.example server/.env  # укажите ключи Infobip
   - cd infra && docker-compose up -d --build
   - GET http://localhost:8080/health -> {"status":"ok"}

2) Android:
   - Откройте android/ в Android Studio
   - Сборка: ./gradlew :app:assembleDebug

3) OTP:
   - POST /v1/auth/otp/request {"phone":"+15551234567"}
   - POST /v1/auth/otp/verify {"phone":"+15551234567","code":"123456"}

Примечание: версии Gradle/AGP/Kotlin/Compose можно обновить автофиксами Android Studio.
