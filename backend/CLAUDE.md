# WAPHafiz Backend — Rules for Claude

Stack: **Golang · Wapgo boilerplate · Fiber v2 · GORM · PostgreSQL · Redis**

---

## Konteks Project

Aplikasi hafalan Quran untuk keluarga (10–20 orang). Fitur utama:
- Auth dengan kode undangan keluarga (bukan open registration)
- Tracker hafalan per ayat (status: belum/sedang/hafal)
- Sistem muraja'ah otomatis: **Sabqi** (hafalan 7 hari terakhir) + **Manzil** (siklus 7 hari seluruh hafalan)
- Proxy + cache Quran API (`alquran.cloud`) di Redis TTL 24 jam
- Dashboard progress per anggota; admin bisa lihat semua

---

## Struktur Direktori

```
cmd/api/main.go              ← dependency wiring + entrypoint
config/                      ← Viper ENV-first config
internal/
  domain/                    ← entities, repository interfaces, service interfaces
  usecase/                   ← business logic
  delivery/http/             ← handlers, middleware, routes
  repository/                ← GORM (db/) dan Redis (redis/) implementations
pkg/
  auth/                      ← JWT middleware
  httpclient/                ← resilient inter-service HTTP client
  messaging/                 ← Kafka/RabbitMQ (belum dipakai v1)
  observability/             ← OpenTelemetry / Elastic APM (disabled v1)
  logger/                    ← zerolog structured logging
  response/                  ← helper Success, Paginated, Error
migrations/                  ← SQL migration files
```

---

## Cara Kerja Wapgo

Scaffold domain baru:
```bash
wapgo make:all <name>        # generate semua layer sekaligus
wapgo make:model <name>
wapgo make:repo <name>
wapgo make:usecase <name>
wapgo make:controller <name>
wapgo make:route <name>
wapgo make:migration <name>
```

Setelah generate, **wajib wire manual** di dua tempat:
1. `cmd/api/main.go` — inject handler + repository
2. `internal/delivery/http/route/router.go` — daftarkan route

---

## Rules Wajib

### Arsitektur

- Ikuti **Clean Architecture**: domain → usecase → delivery. Jangan import delivery dari usecase atau domain.
- Entity ada di `internal/domain/`. Jangan taruh business logic di handler.
- Repository hanya boleh akses DB/Redis — tidak boleh ada logic bisnis di sini.
- Usecase tidak boleh tahu tentang Fiber atau HTTP — gunakan struct request/response biasa.

### Database & Migration

- Semua skema perubahan lewat migration file (`wapgo make:migration <name>`).
- Jangan pakai `AutoMigrate` GORM di production. Hanya untuk local dev jika benar-benar darurat.
- Schema yang perlu dibuat (urutan):
  1. `family_groups` — `id, name, invite_code (unique), created_by, created_at, updated_at`
  2. `users` — tambah field `family_group_id`, `role (admin/member)`, `is_active`
  3. `hafalan_progress` — `id, user_id, surah_number, ayat_start, ayat_end, status, noted_at, created_at, updated_at`
  4. `murajaah_schedule` — `id, user_id, type (sabqi/manzil), surah_number, ayat_start, ayat_end, scheduled_date, completed_at, created_at`
  5. `murajaah_log` — `id, schedule_id, user_id, completed_at, notes`

### Redis

- Cache Quran API response dengan key `quran:surah:{number}` dan `quran:ayah:{ref}:{edition}`, TTL 24 jam.
- Selalu cek cache dulu sebelum hit `alquran.cloud`. Kalau keduanya gagal, return error jelas — jangan crash.
- Jangan simpan JWT di Redis manual; Wapgo sudah handle blacklist token.

### Auth & RBAC

- Register **wajib** pakai `invite_code` yang valid. Tolak kalau kode tidak ada atau tidak aktif.
- JWT sudah di-handle Wapgo. Frontend handle 401 → redirect login.
- Gunakan middleware `RequireRole("admin")` untuk semua endpoint admin.
- Role hanya dua: `admin` dan `member`.

### Hafalan — Business Logic Kritis

- **Cek overlap** sebelum insert hafalan baru. Jika rentang ayat overlap dengan record existing milik user yang sama pada surah yang sama: merge range atau minta konfirmasi — jangan simpan duplikat.
- `ayat_start` harus `<= ayat_end`. Validasi di usecase, bukan hanya di handler.
- Status: `belum` → `sedang` → `hafal`. Bisa diupdate PATCH.

### Muraja'ah — Business Logic Kritis

- **Sabqi**: generate dari hafalan yang `noted_at` dalam 7 hari terakhir.
- **Manzil**: ambil seluruh hafalan user, bagi rata ke 7 hari, rotate berdasarkan `scheduled_date mod 7`.
- Sebelum generate schedule hari ini: cek dulu apakah sudah ada record di `murajaah_schedule` untuk `user_id` + `scheduled_date = today`. Kalau sudah ada, skip — jangan duplikat.
- Kalau user belum punya hafalan sama sekali: skip generate, jangan return error.
- Trigger generate: saat user hit `GET /api/v1/murajaah/today` (lazy generation).

### Response Format

Gunakan helper dari `pkg/response`:
```go
response.Success(c, data)
response.Paginated(c, data, meta)
response.Error(c, code, message)
```

Jangan return raw JSON manual dari handler.

### Logging

- Gunakan `pkg/logger` (zerolog). Jangan pakai `fmt.Println` atau `log.Printf`.
- Log level: `ERROR` untuk unexpected errors, `INFO` untuk lifecycle events, `DEBUG` untuk detail request/response (hanya development).

### Error Handling

- External API (`alquran.cloud`) timeout atau error → return `503` dengan pesan yang jelas ke client, bukan panic.
- DB error → log + return `500`, jangan expose raw SQL error ke client.
- Validation error → return `400` dengan field yang bermasalah.

---

## Environment Variables Penting

```bash
APP_ENV=development        # development | production
APP_PORT=8080
DB_DRIVER=postgres
DB_DSN=...
REDIS_ADDR=localhost:6379
JWT_SECRET=...
OBSERVABILITY_PROVIDER=none   # none | otel | elastic_apm
```

---

## Make Targets

```bash
make run           # jalankan lokal
make docker-up     # start PostgreSQL + Redis
make docker-down   # stop dependencies
make test          # unit test
make test-race     # test + race detector
make coverage      # coverage >= 80%
make lint          # golangci-lint
make check         # lint + sec + test-race + coverage
make tidy          # go mod tidy
```

---

## API Endpoints (Target v1)

### Auth
```
POST /api/v1/auth/register      body: {invite_code, name, email, password}
POST /api/v1/auth/login         body: {email, password}
POST /api/v1/auth/logout
```

### Family (Admin only)
```
GET    /api/v1/family/members
PATCH  /api/v1/family/members/:id/status
POST   /api/v1/family/invite-code/regenerate
```

### Quran (proxy + cache)
```
GET /api/v1/quran/surah
GET /api/v1/quran/surah/:number
GET /api/v1/quran/ayah/:surah/:ayat/audio
```

### Hafalan
```
POST   /api/v1/hafalan
GET    /api/v1/hafalan
PATCH  /api/v1/hafalan/:id
GET    /api/v1/hafalan/summary
GET    /api/v1/admin/hafalan         (admin only)
```

### Muraja'ah
```
GET  /api/v1/murajaah/today
POST /api/v1/murajaah/:id/complete
GET  /api/v1/murajaah/history
```

### Dashboard
```
GET /api/v1/dashboard/me
GET /api/v1/admin/dashboard          (admin only)
```

---

## External API — alquran.cloud

```
GET https://api.alquran.cloud/v1/surah
GET https://api.alquran.cloud/v1/surah/{number}
GET https://api.alquran.cloud/v1/ayah/{ref}/ar.alafasy
```

Taruh HTTP client di `pkg/quranapi/`. Gunakan `pkg/httpclient` yang sudah ada untuk resilience (retry, timeout).

---

## Scope v1 — Jangan Dikerjakan Sekarang

- Voice recognition / koreksi bacaan
- Hafalan hadits
- Push notification
- Kafka / RabbitMQ (sudah ada di boilerplate, tapi jangan aktifkan)
- Observability (set ke `none`)
- Export PDF, statistik jangka panjang
