# WAPHafiz — Project Rules

Aplikasi hafalan Quran untuk keluarga internal (10–20 orang).
Admin memantau progress seluruh anggota; anggota mencatat hafalan dan mengikuti jadwal muraja'ah otomatis.

Baca PRD lengkap di `docs/prd.md` dan dev plan di `docs/devplan.md` sebelum mengerjakan task apapun.

---

## Struktur Monorepo

```
waphafiz/
├── backend/      ← Golang (Wapgo · Fiber v2 · GORM)
├── frontend/     ← Nuxt 3 (Vue 3 · Tailwind · Pinia)
├── design/       ← HTML mockup + design.md (design system reference)
└── docs/         ← prd.md, devplan.md
```

Tiap direktori punya `CLAUDE.md` sendiri dengan rules spesifik.
Rules di bawah ini berlaku untuk **seluruh project**.

---

## Domain & Terminologi

Gunakan istilah ini secara konsisten di kode, variabel, API, dan UI:

| Istilah | Arti |
|---|---|
| `sabak` | Hafalan baru hari ini |
| `sabqi` | Muraja'ah hafalan 7 hari terakhir |
| `manzil` | Siklus muraja'ah seluruh hafalan lama (7 hari) |
| `hafalan_progress` | Record ayat yang sudah dihafal user |
| `murajaah_schedule` | Jadwal muraja'ah harian yang di-generate otomatis |
| `family_group` | Satu keluarga dengan satu invite code |
| `admin` | Kepala keluarga, bisa kelola anggota |
| `member` | Anggota biasa |
| Status hafalan | `belum` · `sedang` · `hafal` |
| Tipe muraja'ah | `sabqi` · `manzil` |

---

## Scope v1 — Frozen

Jangan tambah fitur di luar ini:

**Yang dikerjakan:**
- Auth dengan invite code keluarga
- Tracker hafalan per ayat (sabak)
- Sistem muraja'ah otomatis (sabqi + manzil)
- Proxy + cache Quran API (alquran.cloud)
- Dashboard progress per anggota
- Admin: kelola anggota, lihat progress semua

**Yang TIDAK dikerjakan di v1:**
- Voice recognition / koreksi bacaan
- Hafalan hadits
- Push notification
- Mobile app native
- Monetisasi, fitur sosial, komunitas luar keluarga
- Export PDF, grafik statistik jangka panjang

---

## Aturan Umum

### Bahasa

- **Kode (variabel, fungsi, file):** Bahasa Inggris
- **UI (label, pesan, teks):** Bahasa Indonesia
- **Komentar kode:** Boleh Indonesia atau Inggris, pilih satu dan konsisten per file

### Tidak Boleh

- Jangan commit secret, JWT secret, password, API key ke repo
- Jangan hard-code URL production di kode — selalu pakai environment variable
- Jangan bypass validasi dengan alasan "nanti diperbaiki"
- Jangan tambah dependency baru tanpa alasan yang jelas

### Commit Convention

Format: `type(scope): pesan singkat`

```
feat(auth): register dengan invite code keluarga
fix(murajaah): skip generate kalau hafalan kosong
chore(deps): update fiber v2.52.5
```

Type: `feat` · `fix` · `refactor` · `test` · `chore` · `docs`
Scope: `auth` · `hafalan` · `murajaah` · `quran` · `dashboard` · `admin` · `frontend` · `backend`

---

## Data Flow

```
[Nuxt 3 Frontend]
      │  REST API JSON (JWT Bearer)
      ▼
[Golang Backend — port 8080]
      │              │
[PostgreSQL]      [Redis]
                    │
              [alquran.cloud]  ← cache 24 jam TTL
```

- Frontend tidak boleh akses DB atau alquran.cloud langsung — semua lewat backend
- Backend adalah satu-satunya yang boleh menyentuh Redis dan PostgreSQL

---

## API Contract

- Base URL: `http://localhost:8080/api/v1` (dev) — via env di frontend
- Auth: `Authorization: Bearer <jwt>`
- Response envelope selalu:
  ```json
  { "success": true, "data": {...} }
  { "success": false, "message": "...", "errors": {...} }
  ```
- Pagination: `{ "data": [...], "meta": { "page", "limit", "total" } }`
- Tanggal: ISO 8601 string (`2026-06-14T07:00:00Z`)

---

## Design System

Referensi lengkap di `design/design.md`. Ringkasan:

- **Primary color:** `#1D9E75` (green)
- **Font:** Inter (UI) + Amiri (teks Arab, direction rtl)
- **Icon library:** Tabler Icons (`@tabler/icons-vue` di Nuxt)
- **Border radius:** sm=6px · md=10px · lg=14px · xl=20px
- Mobile-first: sidebar desktop → bottom nav mobile

---

## Testing

- Backend: unit test usecase; integration test flow kritis (register → login → hafalan → muraja'ah)
- Frontend: test komponen kritis dengan Vitest; manual test di mobile browser (Safari iOS, Chrome Android)
- Coverage backend minimal 80% (`make coverage`)
- Jangan mock DB di integration test backend

---

## Docker

```bash
# Jalankan semua dependencies
cd backend && make docker-up

# Health check
curl http://localhost:8080/health
```

`docker-compose.yml` ada di `backend/`. Untuk production pakai `backend/deploy/`.
