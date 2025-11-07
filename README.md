# Kent — skeleton

Готовый скелет мессенджера Kent:

- Android (Kotlin/Compose/Hilt), package: com.kent.app
- Go API (Gin + Redis)
- Docker Compose (Redis + API)
- CI (GitHub Actions, GitLab CI) + черновики документов для Google Play

## Ветки и релизы

- main — стабильные релизы, теги X.Y.Z. Merge только через PR после успешных пайплайнов и ревью.
- develop — интеграционная ветка. Feature/hotfix ветки ответвляются от неё и попадают через PR.
- eature/* — фичи. По завершении — PR в develop.
- hotfix/* — срочные фиксы для main, после merge — обратно в develop.

## CI/CD

### GitHub Actions

Workflow ndroid-ci.yml и go-ci.yml запускаются для веток main и develop.

Переменные (Actions → Secrets and variables → Actions):

- INFOBIP_BASE_URL
- INFOBIP_API_KEY
- INFOBIP_FROM
- OTP_SECRET
- OTP_TTL_SECONDS (300)

### GitLab CI

.gitlab-ci.yml содержит jobs go_build и ndroid_debug (правила main/develop).

Переменные (Settings → CI/CD → Variables): те же значения, Mask для чувствительных данных.

## Запуск dev-стенда

`ash
cp server/.env.example server/.env
# заполните InfoBip и OTP_SECRET
cd infra
# с Redis и API в Docker (на локальном Git нельзя использовать Postgres/NATS/MinIO по умолчанию)
docker-compose up -d --build
curl http://localhost:8080/health
`

## Android проект

Открыть каталог ndroid/ в Android Studio:

`ash
./gradlew :app:assembleDebug
`

## Секции docs/

- PRIVACY_POLICY.md
- TERMS.md
- PLAY_LISTING.md
- DATA_SAFETY.md
- API.md

## Лицензия

Apache License 2.0 (LICENSE).

## Contributing

Правила PR/веток и чек-листы описаны в CONTRIBUTING.md.
