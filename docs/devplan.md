# Dev Plan — Aplikasi Hafalan Quran Keluarga

**Versi:** 1.0  
**Stack:** Wapgo (Golang) + Nuxt 3 + PostgreSQL + Redis  
**Target:** Internal keluarga, 10–20 user

---

## Status Legend

- `[ ]` Belum dikerjakan
- `[~]` Sedang dikerjakan
- `[x]` Selesai

---

## Phase 0 — Setup & Fondasi

### Backend
- [x] Init project wapgo: `wapgo new quran-family-app --module github.com/me/quran-family-app`
- [x] Setup `.env` — DB, Redis, JWT secret, APP_ENV
- [x] `make docker-up` — pastikan PostgreSQL + Redis jalan
- [x] Buat database `quran_family_db`
- [x] Verifikasi `curl http://localhost:8080/health` OK

### Frontend
- [x] Init Nuxt 3: `npx nuxi init quran-family-web`
- [x] Install dependencies: `@nuxtjs/tailwindcss`, `pinia`, `@vueuse/nuxt`
- [x] Setup `runtimeConfig` untuk base URL backend
- [x] Setup Pinia store skeleton (auth, hafalan, muraja'ah)
- [x] Buat layout dasar: sidebar + main content area

---

## Phase 1 — Auth & Family Group

### Backend
- [x] Generate domain user (sudah ada di wapgo, sesuaikan)
- [x] Tambah field `family_group_id` dan `role` ke tabel `users`
- [x] Buat migration: tabel `family_groups`
  ```sql
  id, name, invite_code (unique), created_by, created_at, updated_at
  ```
- [x] Endpoint: `POST /api/v1/auth/register` — wajib pakai `invite_code`
- [x] Endpoint: `POST /api/v1/auth/login` (sudah ada di wapgo)
- [x] Endpoint: `POST /api/v1/auth/logout` (sudah ada di wapgo)
- [x] Endpoint: `GET /api/v1/family/members` — admin only, list semua anggota
- [x] Endpoint: `PATCH /api/v1/family/members/:id/status` — admin, aktif/nonaktif anggota
- [x] Endpoint: `POST /api/v1/family/invite-code/regenerate` — admin, buat ulang kode undangan
- [x] Middleware RBAC: `RequireRole("admin")` untuk endpoint admin
- [x] Seed data: 1 family group + 1 admin user untuk testing

### Frontend
- [x] Halaman `/login` — form email + password
- [x] Halaman `/register` — form nama, email, password, kode undangan
- [x] Auth store (Pinia): simpan JWT, refresh token, user info
- [x] Route guard: redirect ke `/login` kalau belum auth
- [x] Halaman `/admin/members` — tabel daftar anggota (admin only)
- [x] Tombol aktif/nonaktif anggota

---

## Phase 2 — Integrasi Data Quran

### Backend
- [x] Buat `pkg/quranapi` — HTTP client ke alquran.cloud
- [x] Fungsi: `GetAllSurah()` → list 114 surah + metadata
- [x] Fungsi: `GetSurahDetail(number int)` → ayat lengkap per surah
- [x] Fungsi: `GetAyahAudio(ref string, edition string)` → URL audio per ayat
- [x] Cache semua response di Redis, TTL 24 jam
- [x] Endpoint: `GET /api/v1/quran/surah` — list semua surah
- [x] Endpoint: `GET /api/v1/quran/surah/:number` — detail surah + ayat
- [x] Endpoint: `GET /api/v1/quran/ayah/:surah/:ayat/audio` — URL audio ayat
- [x] Error handling: kalau alquran.cloud timeout, return cached data atau 503 dengan pesan jelas

### Frontend
- [x] Halaman `/quran` — daftar 114 surah dengan info (nama, jumlah ayat, juz)
- [x] Halaman `/quran/:surah` — tampilkan ayat per surah
- [x] Komponen `AyahCard` — tampil nomor ayat, teks Arab, terjemahan
- [x] Komponen `AudioPlayer` — play/pause audio per ayat (HTML5 audio)
- [x] Loading state saat fetch data Quran

---

## Phase 3 — Tracker Hafalan (Sabak)

### Backend
- [x] Buat migration: tabel `hafalan_progress`
  ```sql
  id, user_id, surah_number, ayat_start, ayat_end,
  status (belum/sedang/hafal), noted_at, created_at, updated_at
  ```
- [x] `wapgo make:all hafalan`
- [x] Endpoint: `POST /api/v1/hafalan` — catat hafalan baru
  - Validasi: ayat_start <= ayat_end, surah valid
  - Tidak boleh overlap dengan hafalan yang sudah ada (merge atau update)
- [x] Endpoint: `GET /api/v1/hafalan` — list hafalan milik user sendiri
- [x] Endpoint: `PATCH /api/v1/hafalan/:id` — update status ayat
- [x] Endpoint: `DELETE /api/v1/hafalan/:id` — hapus hafalan milik sendiri
- [x] Endpoint: `GET /api/v1/hafalan/summary` — ringkasan: berapa juz/surah sudah hafal
- [x] Endpoint: `GET /api/v1/admin/hafalan` — admin: progress semua anggota (admin only)
- [x] Endpoint: `GET /api/v1/admin/hafalan/member/:userID` — admin: list hafalan per member
- [x] Endpoint: `PATCH /api/v1/admin/hafalan/:id` — admin: edit status hafalan member
- [x] Endpoint: `DELETE /api/v1/admin/hafalan/:id` — admin: hapus hafalan member
- [x] Logic: hitung persentase hafalan per surah dan per juz

### Frontend
- [x] Halaman `/hafalan` — daftar progress hafalan saya
  - Tampilkan per surah: surah apa, ayat berapa, status
  - Filter: tampilkan yang sudah hafal / sedang / belum
  - Tombol toggle status (Hafal/Sedang/Belum) per card
  - Tombol hapus per card dengan konfirmasi
- [x] Tombol "Tandai Hafal" dari halaman `/quran/:surah`
  - Dialog: pilih ayat_start dan ayat_end
  - Submit → hit POST /api/v1/hafalan
- [x] Progress bar per surah dan per juz
- [x] Halaman `/admin/progress` — progress semua anggota + modal kelola hafalan per member

---

## Phase 4 — Sistem Muraja'ah (Sabqi & Manzil)

### Backend
- [x] Buat migration: tabel `murajaah_schedule`
  ```sql
  id, user_id, type (sabqi/manzil), surah_number, ayat_start, ayat_end,
  scheduled_date, completed_at, created_at
  ```
- [x] Buat migration: tabel `murajaah_log`
  ```sql
  id, schedule_id, user_id, completed_at, notes
  ```
- [x] `wapgo make:all murajaah`
- [x] Logic Sabqi: setiap hari generate jadwal dari hafalan 7 hari terakhir
- [x] Logic Manzil: bagi seluruh hafalan ke siklus 7 hari, rotate setiap hari
- [x] Cron job atau trigger saat login: generate schedule untuk hari ini kalau belum ada
- [x] Endpoint: `GET /api/v1/murajaah/today` — jadwal muraja'ah hari ini (sabqi + manzil)
- [x] Endpoint: `POST /api/v1/murajaah/:id/complete` — tandai muraja'ah selesai
- [x] Endpoint: `GET /api/v1/murajaah/history` — riwayat muraja'ah

### Frontend
- [x] Halaman `/murajaah` — jadwal hari ini
  - Tampilkan list sabqi dan manzil hari ini
  - Tiap item: nama surah, rentang ayat, tombol "Selesai"
- [x] Banner reminder di dashboard: "Kamu belum muraja'ah hari ini"
- [ ] Halaman `/murajaah/history` — kalender / list riwayat

---

## Phase 5 — Dashboard & Polish

### Backend
- [x] Endpoint: `GET /api/v1/dashboard/me` — summary pribadi
  - Total hafalan (ayat/surah/juz)
  - Streak hari berturut-turut
  - Status muraja'ah hari ini
- [x] Endpoint: `GET /api/v1/admin/dashboard` — summary semua anggota
  - Siapa yang aktif hari ini
  - Ranking berdasarkan total hafalan

### Frontend
- [x] Halaman `/dashboard` — homepage setelah login
  - Stats: total hafalan, muraja'ah hari ini, progress hafalan per surah (grouped)
  - Jadwal muraja'ah hari ini (ringkasan 3 item)
  - Banner reminder kalau muraja'ah belum selesai
- [x] Halaman `/admin/dashboard` — versi admin
  - Stats total anggota, aktif hari ini, total ayat hafal
- [x] Halaman `/murajaah/history` — riwayat muraja'ah dengan filter tipe
- [~] Responsive: pastikan semua halaman nyaman di mobile browser
- [x] Empty state yang informatif (kalau belum ada hafalan, tampilkan panduan mulai)
- [x] Error handling global: toast notification untuk error API

---

---

## Phase 5b — Komponen UI & Infrastruktur Frontend

Item ini tidak ada di plan awal tapi sudah dikerjakan:

### Komponen UI (semua di `components/ui/`)
- [x] `AppBadge` — badge status dengan variant (green, amber, gray, blue)
- [x] `AppProgressBar` — progress bar dengan warna dan tinggi configurable
- [x] `AppAvatar` — avatar initial nama dengan warna otomatis
- [x] `AppToast` — komponen toast notifikasi

### Komponen Domain
- [x] `HafalanModal` — modal tambah hafalan: pilih surah, ayat_start, ayat_end, status; validasi overlap + limit ayat
- [x] `MurajaahCard` — card jadwal muraja'ah harian
- [x] `QuranAyahCard` — card ayat dengan teks Arab, terjemahan, tombol audio
- [x] `QuranAudioPlayer` — audio player per ayat (HTML5, play/pause/repeat)

### Layout
- [x] `layouts/default.vue` — sidebar desktop + bottom nav mobile
- [x] `layouts/auth.vue` — layout centered untuk login/register

### Infrastruktur
- [x] `utils/api.ts` — wrapper `$fetch` dengan base URL dari runtimeConfig + Authorization header
- [x] `utils/surahNames.ts` — mapping nomor surah → nama Indonesia (114 surah)
- [x] Nuxt config: `pathPrefix: false` agar komponen di-resolve by filename, bukan path prefix
- [x] CORS middleware backend: tambah `PATCH` ke `AllowMethods`
- [x] Audio repeat per hafalan di halaman `/hafalan` — play loop ayat_start sampai ayat_end

---

## Phase 6 — Testing & Deploy

### Testing
- [ ] Unit test backend: usecase hafalan, logic sabqi/manzil
- [ ] Integration test: flow register → login → catat hafalan → get muraja'ah
- [ ] Manual testing dengan anggota keluarga (minimal 3 orang)
- [ ] Test di mobile browser (Safari iOS, Chrome Android)

### Deploy
- [ ] Setup `docker-compose.prod.yml` (backend + postgres + redis)
- [ ] Build Nuxt: `nuxi build` + setup static hosting atau Node server
- [ ] Setup environment variables production
- [ ] Deploy ke VPS (bisa pakai yang paling murah, traffic kecil)
- [ ] Setup domain atau subdomain
- [ ] Test end-to-end di production

---

## Backlog v2 (Jangan Dikerjakan Sekarang)

- [ ] Voice recognition untuk koreksi bacaan (Whisper API)
- [ ] Hafalan hadits
- [ ] Push notification (butuh mobile app)
- [ ] React Native mobile app
- [ ] Export progress ke PDF
- [ ] Statistik hafalan jangka panjang (grafik bulanan/tahunan)

---

## Catatan Teknis Penting

**Quran API fallback:**
Selalu cek cache Redis sebelum hit alquran.cloud. Kalau keduanya gagal, return error yang jelas — jangan crash.

**Overlap hafalan:**
Saat user input hafalan baru, cek apakah rentang ayat overlap dengan record yang sudah ada. Kalau overlap, merge atau minta user konfirmasi — jangan simpan duplikat.

**Generate muraja'ah:**
Jangan generate ulang schedule kalau sudah ada untuk hari itu. Cek dulu sebelum insert. Kalau belum ada hafalan sama sekali, skip — jangan error.

**JWT:**
Wapgo sudah handle refresh token + blacklist. Pastikan frontend handle 401 dengan redirect ke login, bukan infinite loop retry.
