# PRD — Aplikasi Hafalan Quran Keluarga

**Versi:** 1.0  
**Author:** Abdullah Prasetio  
**Status:** Draft  
**Tanggal:** Juni 2026

---

## 1. Latar Belakang

Aplikasi hafalan Quran yang ada di pasar (Tarteel, Quran Companion) mayoritas berbayar dan dirancang untuk pengguna umum. Tidak ada solusi yang dirancang khusus untuk kebutuhan keluarga — dimana satu admin bisa memantau progress hafalan seluruh anggota keluarga secara bersama.

Aplikasi ini dibangun untuk keluarga sendiri (10–20 orang), dengan fokus pada tracking progress hafalan dan sistem muraja'ah yang sistematis mengikuti metode pesantren (Sabak–Sabqi–Manzil).

---

## 2. Tujuan

- Membantu anggota keluarga menghafal Quran secara terstruktur
- Menyediakan sistem muraja'ah otomatis agar hafalan tidak mudah hilang
- Memungkinkan kepala keluarga / admin memantau progress semua anggota
- Gratis, privat, dan hanya untuk keluarga

---

## 3. Non-Goals (Tidak Dikerjakan di v1)

- Voice recognition / koreksi bacaan otomatis
- Hafalan hadits (ditunda ke v2)
- Fitur sosial / komunitas luar keluarga
- Mobile app native (web-first dulu)
- Monetisasi

---

## 4. Pengguna

| Peran | Deskripsi |
|---|---|
| **Admin (Kepala Keluarga)** | Mengelola anggota, melihat semua progress, mengatur grup |
| **Anggota** | Mencatat hafalan harian, mengikuti jadwal muraja'ah |

Target pengguna awal: 10–20 orang dalam satu keluarga besar.

---

## 5. Fitur Utama

### 5.1 Autentikasi & Manajemen Keluarga

- Register dengan kode undangan keluarga (bukan open registration)
- Login dengan email + password
- Role: `admin` dan `member`
- Admin bisa menambah / menonaktifkan anggota

### 5.2 Tracker Hafalan (Sabak)

- Tandai ayat/surah yang sudah dihafal hari ini
- Status per ayat: `belum` / `sedang` / `hafal`
- Input bisa per ayat atau per rentang ayat (misal: Al-Baqarah 1–5)
- Riwayat hafalan harian

### 5.3 Sistem Muraja'ah (Sabqi & Manzil)

- **Sabqi:** Muraja'ah hafalan 7 hari terakhir — dijadwalkan otomatis tiap hari
- **Manzil:** Seluruh hafalan yang sudah ada dibagi rata ke dalam siklus 7 hari
- Notifikasi / reminder muraja'ah harian (via in-app banner, bukan push notification dulu)
- Tandai muraja'ah sebagai selesai

### 5.4 Audio Quran

- Playback audio per ayat dari sumber eksternal (alquran.cloud)
- Pilihan qori (minimal 1–2 pilihan populer)
- Cache audio di sisi client untuk menghemat bandwidth

### 5.5 Dashboard Progress

- Progress hafalan per anggota (berapa juz/surah sudah hafal)
- Grafik aktivitas hafalan mingguan
- Admin view: progress semua anggota dalam satu halaman
- Member view: progress pribadi + posisi muraja'ah hari ini

---

## 6. User Stories Prioritas Tinggi

```
Sebagai anggota, saya bisa mencatat ayat yang baru saya hafal hari ini
agar progress saya tersimpan dan terlacak.

Sebagai anggota, saya mendapat jadwal muraja'ah otomatis setiap hari
agar saya tidak lupa mengulang hafalan lama.

Sebagai admin, saya bisa melihat progress hafalan semua anggota keluarga
agar saya tahu siapa yang aktif dan siapa yang perlu didorong.

Sebagai anggota, saya bisa mendengarkan audio ayat
agar saya bisa menghafal dengan talaqqi yang benar.
```

---

## 7. Arsitektur Sistem

```
[Nuxt 3 Frontend]
        │
        │ REST API (JSON)
        ▼
[Wapgo — Golang Backend]
        │
   ┌────┴────┐
   │         │
[PostgreSQL] [Redis]
              (cache Quran API + session)
        │
        ▼
[alquran.cloud API] ← sumber data Quran & audio
```

---

## 8. Stack Teknologi

| Layer | Teknologi |
|---|---|
| Backend | Golang (Wapgo boilerplate) — Fiber v2, GORM |
| Database | PostgreSQL |
| Cache | Redis |
| Auth | JWT HS256 + RBAC (sudah ada di Wapgo) |
| Frontend | Nuxt 3 (Vue 3) |
| Data Quran | alquran.cloud API (gratis, open) |
| Deployment | Docker + docker-compose |

---

## 9. Data Model (High-Level)

```
family_groups
  id, name, invite_code, created_by, created_at

users
  id, name, email, password_hash, family_group_id,
  role (admin/member), is_active, created_at

hafalan_progress
  id, user_id, surah_number, ayat_start, ayat_end,
  status (belum/sedang/hafal), noted_at, created_at

muraja'ah_schedule
  id, user_id, type (sabqi/manzil),
  surah_number, ayat_start, ayat_end,
  scheduled_date, completed_at

muraja'ah_log
  id, schedule_id, user_id, completed_at, notes
```

---

## 10. API Quran Eksternal

Menggunakan **alquran.cloud API v1** — gratis, tidak perlu API key.

```
GET https://api.alquran.cloud/v1/surah          → list semua surah
GET https://api.alquran.cloud/v1/surah/{nomor}  → detail + ayat
GET https://api.alquran.cloud/v1/ayah/{ref}/ar.alafasy → audio per ayat
```

Response di-cache di Redis dengan TTL 24 jam untuk menghindari request berulang.

---

## 11. Kriteria Sukses (v1)

- Seluruh anggota keluarga bisa login dan mencatat hafalan tanpa error
- Jadwal muraja'ah Sabqi & Manzil ter-generate otomatis setelah ada hafalan
- Admin bisa melihat dashboard progress semua anggota
- Audio ayat bisa diputar di browser (desktop & mobile)
- Performa: halaman utama load < 2 detik

---

## 12. Timeline Target

| Milestone | Estimasi |
|---|---|
| Setup project + auth + family group | Minggu 1 |
| Integrasi Quran API + cache Redis | Minggu 1–2 |
| Domain hafalan — CRUD progress | Minggu 2 |
| Sistem muraja'ah otomatis | Minggu 3 |
| Frontend Nuxt — semua halaman | Minggu 3–4 |
| Testing internal keluarga | Minggu 5 |

---

## 13. Risiko

| Risiko | Mitigasi |
|---|---|
| alquran.cloud API down | Cache Redis 24 jam; fallback pesan "audio tidak tersedia" |
| Anggota tidak konsisten input hafalan | Reminder in-app harian; dashboard admin untuk dorongan |
| Scope creep (minta fitur baru terus) | Freeze v1 scope; catat permintaan untuk v2 |
