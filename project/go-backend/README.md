# MasterFabric Go Backend

MasterFabric Academy agentic AI development programı için geliştirilen backend.
Gin + GORM (Postgres) + JWT auth ile yazıldı. "Daily English Writing Coach"
uygulamasının backend'idir: kullanıcılar günün konusu hakkında İngilizce essay
yazar, Ollama (gemma2:2b) essay'i rubriğe göre puanlar ve hata listesi çıkarır.

## Klasör yapısı

```
cmd/api/main.go              -> giriş noktası, router kurulumu
internal/config/              -> ortam değişkeni yönetimi
internal/database/             -> Postgres bağlantısı + auto-migration + topic seed
internal/models/               -> User, Topic, Essay, EssayScore
internal/handlers/             -> auth.go, config.go, topics.go, writing.go, common.go
internal/services/             -> essay_scoring_prompt.go, scoring.go (Ollama skorlama)
internal/llmclient/            -> Ollama HTTP client
internal/middleware/           -> JWT auth middleware, CORS
internal/metrics/              -> Prometheus metrikleri (mlcmon_writing_*)
internal/utils/                -> JWT üretme/doğrulama, bcrypt şifreleme
```

## API Endpoint'leri

- **Auth**: `/api/auth/*` (register, login, logout, refresh, forgot/reset password, verify-email, me)
- **Config**: `GET /api/config`
- **Topics**: `GET /api/topics/daily`, `GET /api/topics/random`, `POST /api/topics/custom`
- **Essays**: `POST /api/essays`, `GET /api/essays`, `GET /api/essays/:id`
- **Stats**: `GET /api/stats/dashboard`, `GET /api/stats/me`, `GET /api/stats/streak`
- **Health**: `GET /api/health`, `GET /api/version`
- **Metrics**: `GET /metrics` (Prometheus)

## Yerel kurulum

1. Go 1.22+ kurulu olmalı.
2. Postgres'i yerelde çalıştır (veya Docker: `docker run -e POSTGRES_PASSWORD=pass -p 5432:5432 postgres`).
3. `.env.example` dosyasını `.env` olarak kopyala ve değerleri doldur:
   ```bash
   cp .env.example .env
   ```
4. Bağımlılıkları indir:
   ```bash
   go mod tidy
   ```
5. Sunucuyu başlat:
   ```bash
   go run cmd/api/main.go
   ```
6. Sağlık kontrolü:
   ```bash
   curl http://localhost:8080/api/health
   ```

## Render'a deploy

1. Render dashboard'da **New > Web Service** seç, bu repoyu bağla.
2. Build command: `go build -o app cmd/api/main.go`
3. Start command: `./app`
4. Render'da bir **Postgres** instance oluştur, "Internal Database URL"
   değerini `DATABASE_URL` environment variable olarak Web Service'e ekle.
5. `JWT_SECRET` ve `ALLOWED_ORIGIN` (Vercel frontend URL'in) değerlerini de
   environment variable olarak ekle.
6. Render, `/api/health` endpoint'ini health check olarak kullanabilir
   (Settings > Health Check Path).

## Notlar

- Uygulama ilk açılışta `topics` tablosuna 60 sistem konusu ekler
  (`database.SeedTopics`).
- Essay skorlama Ollama üzerinden yapılır; `OLLAMA_URL` ve `OLLAMA_MODEL`
  environment variable'ları ile yapılandırılır.
- Stateless JWT kullanıldığı için `logout` endpoint'i şu an sadece 200
  döner; gerçek bir token iptali istersen bir refresh-token blacklist
  tablosu eklemek gerekir.
- CORS sadece `ALLOWED_ORIGIN` değerine izin verir — Vercel'e deploy
  ettiğinde bu değeri güncellemeyi unutma.