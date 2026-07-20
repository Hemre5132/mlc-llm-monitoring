# MasterFabric Go Backend

MasterFabric Academy agentic AI development programı için geliştirilen backend.
Gin + GORM (Postgres) + JWT auth ile yazıldı. 22 endpoint içerir (Auth 8,
Config 2, Web MLC-LLM 8, Common Services 4).

## Klasör yapısı

```
cmd/api/main.go              -> giriş noktası, router kurulumu
internal/config/              -> ortam değişkeni yönetimi
internal/database/             -> Postgres bağlantısı + auto-migration
internal/models/               -> User, LLMSession, LLMMessage, LLMScore
internal/handlers/             -> auth.go, config.go, llm.go, common.go
internal/middleware/           -> JWT auth middleware, CORS
internal/utils/                -> JWT üretme/doğrulama, bcrypt şifreleme
```

## Yerel kurulum

1. Go 1.22+ kurulu olmalı.
2. Postgres'i yerelde çalıştır (veya Docker: `docker run -e POSTGRES_PASSWORD=pass -p 5432:5432 postgres`).
3. `.env.example` dosyasını `.env` olarak kopyala ve değerleri doldur:
   ```bash
   cp .env.example .env
   ```
4. Bağımlılıkları indir (bu adım internet gerektirir, bu sandbox'ta
   proxy.golang.org'a erişim kısıtlı olduğu için burada çalıştırılamadı —
   kendi makinende veya Claude Code ortamında çalıştır):
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

- `internal/handlers/llm.go` içindeki `CreateScore` fonksiyonunda
  Deci.Scoring için şimdilik placeholder bir hesaplama var (`TODO` ile
  işaretli) — gerçek skorlama kriterleri ayrı bir adımda eklenecek.
- Stateless JWT kullanıldığı için `logout` endpoint'i şu an sadece 200
  döner; gerçek bir token iptali istersen bir refresh-token blacklist
  tablosu eklemek gerekir.
- CORS sadece `ALLOWED_ORIGIN` değerine izin verir — Vercel'e deploy
  ettiğinde bu değeri güncellemeyi unutma.
