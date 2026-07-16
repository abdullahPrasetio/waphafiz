# WAPHafiz

Aplikasi hafalan Qur'an untuk keluarga internal (10–20 orang). Admin memantau progress seluruh anggota; anggota mencatat hafalan dan mengikuti jadwal muraja'ah otomatis.

Baca `docs/prd.md` dan `docs/devplan.md` untuk detail produk. Rules per-direktori ada di `CLAUDE.md` masing-masing (`backend/`, `frontend/`, root).

## Struktur

```
waphafiz/
├── backend/      ← Golang (Fiber v2 · GORM · PostgreSQL · Redis)
├── frontend/     ← Nuxt 3 (Vue 3 · Tailwind · Pinia)
├── design/       ← HTML mockup + design system
└── docs/         ← PRD, dev plan
```

## Coba build lokal

### 1. Backend

```bash
cp backend/.env.example backend/.env
make deps-up          # start Postgres + Redis (docker-compose)
make backend-run       # jalankan di http://localhost:8080
```

Cek:

```bash
curl http://localhost:8080/health
```

Test:

```bash
make backend-test     # go test -race, termasuk integration test (butuh deps-up)
```

### 2. Frontend

```bash
cp frontend/.env.example frontend/.env
make frontend-install
make frontend-dev      # jalankan di http://localhost:3000
```

Pastikan backend (`make backend-run`) sudah jalan di port 8080 sebelum buka frontend.

### 3. Build Docker image (native platform, untuk tes lokal)

```bash
make build-backend     # image: abdullahprasetio/waphafiz-backend:local
make build-frontend    # image: abdullahprasetio/waphafiz-frontend:local
make build             # keduanya sekaligus
```

Frontend butuh `NUXT_PUBLIC_API_BASE` di-bake saat build kalau backend tidak di `localhost:8080`:

```bash
make build-frontend NUXT_PUBLIC_API_BASE=http://192.168.1.100:8080/api/v1
```

### 4. Full stack via docker-compose

```bash
cp .env.prod.example .env.prod   # edit STB_IP, DB_PASSWORD, JWT_SECRET
make up                          # build + start semua service
make logs                        # tail log
make down                        # stop
```

## CI/CD

Pipeline GitHub Actions ada di [`.github/workflows/ci-cd.yml`](.github/workflows/ci-cd.yml):

- **PR / push apapun**: test backend (Go, dengan Postgres real) + build frontend.
- **Push ke `main`**: build & push image multi-platform (`linux/amd64`, `linux/arm64`) ke Docker Hub dengan tag `latest`.
- **Push tag `v*`**: push image dengan tag versi semver.

Secret yang harus di-set di GitHub (Settings → Secrets and variables → Actions):

| Nama | Tipe | Keterangan |
|---|---|---|
| `DOCKERHUB_USERNAME` | Secret | username Docker Hub |
| `DOCKERHUB_TOKEN` | Secret | access token Docker Hub (bukan password) |
| `NUXT_PUBLIC_API_BASE` | Variable | URL API production, di-bake ke bundle frontend (opsional) |
