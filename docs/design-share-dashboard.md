# Rancangan: Share Dashboard (Link Pantau Read-Only)

> **Status:** DRAFT — menunggu review
> **Tanggal:** 2026-07-20
> **Konteks:** Opsi C dari diskusi multi-group (lihat `design-multi-group.md` §pilihan). Kebutuhan: pihak luar keluarga (guru ngaji, kakek-nenek) bisa memantau progress hafalan seorang anggota **tanpa akun**, lewat link read-only yang bisa dicabut. Skema v1 tidak berubah — fitur ini murni tambahan.

---

## 1. Konsep

```
[Ayah/anggota]                          [Guru ngaji — tanpa akun]
     │ generate link share                      │
     ▼                                          ▼
POST /shares ──► token acak 256-bit ──► https://app/s/{token}
     │              (disimpan HASH-nya)         │
     │                                          ▼
     └─ bisa revoke kapanpun          GET /public/shares/{token}
                                       → dashboard read-only 1 anggota
                                         (tanpa email, tanpa data sensitif)
```

Prinsip:

1. **Read-only mutlak.** Pemegang link hanya bisa melihat; tidak ada satupun endpoint tulis yang menerima share token.
2. **Satu link = satu anggota.** Link memantau dashboard satu orang, bukan seluruh keluarga.
3. **Bisa dicabut kapanpun** oleh pembuatnya (atau admin), dan bisa kedaluwarsa otomatis.
4. **Minim bocoran data.** Payload publik tidak memuat email, user_id asli, atau data anggota lain.

---

## 2. Model Data

Satu tabel baru — tidak ada perubahan pada tabel existing.

### 2.1 Tabel `dashboard_shares`

```sql
CREATE TABLE dashboard_shares (
    id             UUID PRIMARY KEY,
    user_id        UUID         NOT NULL REFERENCES users(id),  -- dashboard siapa
    created_by     UUID         NOT NULL REFERENCES users(id),  -- siapa yang membuat
    token_hash     VARCHAR(64)  NOT NULL UNIQUE,                -- SHA-256 hex dari token
    label          VARCHAR(100) NOT NULL,                       -- "Ustadz Ahmad", "Nenek"
    expires_at     TIMESTAMPTZ,                                 -- NULL = tidak kedaluwarsa
    revoked_at     TIMESTAMPTZ,                                 -- NULL = masih aktif
    last_viewed_at TIMESTAMPTZ,
    view_count     INT          NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_dashboard_shares_user ON dashboard_shares(user_id);
```

Migration: `005_create_dashboard_shares.sql`.

Entity Go (`internal/domain/entity/dashboard_share.go`):

```go
type DashboardShare struct {
    ID           uuid.UUID  `gorm:"type:uuid;primaryKey;not null"  json:"id"`
    UserID       uuid.UUID  `gorm:"type:uuid;not null;index"       json:"user_id"`
    CreatedBy    uuid.UUID  `gorm:"type:uuid;not null"             json:"created_by"`
    TokenHash    string     `gorm:"size:64;uniqueIndex;not null"   json:"-"`
    Label        string     `gorm:"size:100;not null"              json:"label"`
    ExpiresAt    *time.Time `                                      json:"expires_at"`
    RevokedAt    *time.Time `                                      json:"revoked_at"`
    LastViewedAt *time.Time `                                      json:"last_viewed_at"`
    ViewCount    int        `gorm:"not null;default:0"             json:"view_count"`
    CreatedAt    time.Time  `                                      json:"created_at"`
    UpdatedAt    time.Time  `                                      json:"updated_at"`

    User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

Catatan: **tanpa soft delete** — revoke sudah dimodelkan eksplisit lewat `revoked_at`, dan record dipertahankan sebagai audit trail (kapan dibuat, berapa kali dilihat).

### 2.2 Keamanan token — aturan wajib

- **Generate:** 32 byte dari `crypto/rand`, encode `base64.RawURLEncoding` → string ±43 karakter, URL-safe. **Bukan** UUID (UUID v4 cuma 122 bit dan sering bocor di log dengan pola dikenali) dan **bukan** angka/kode pendek.
- **Simpan hash-nya saja** (`SHA-256` hex, 64 char). Token asli hanya muncul **sekali** di response saat pembuatan — persis seperti perlakuan GitHub personal access token. Kalau DB bocor, token tidak bisa dipakai. Lookup: hash input → cari di `token_hash` (O(1) via unique index).
- **Response seragam untuk semua kegagalan:** token tidak ada / salah / revoked / expired semuanya → `404 "Link tidak ditemukan atau sudah tidak berlaku"`. Jangan bedakan pesannya — membedakan = memberi tahu penebak bahwa tokennya "hampir benar".
- **Rate limiting** endpoint publik: Fiber `limiter` middleware, per-IP, mis. 30 req/menit. Dengan token 256-bit brute force memang mustahil secara matematis, tapi rate limit tetap dipasang untuk mencegah abuse/scraping.
- Halaman publik frontend mengirim `<meta name="robots" content="noindex">` — link share tidak boleh terindeks mesin pencari.

---

## 3. Aturan Bisnis

| Aturan | Nilai | Alasan |
|---|---|---|
| Siapa boleh membuat share | User untuk dirinya sendiri; **admin untuk anggota manapun** | Anak kecil tidak mengoperasikan aplikasi sendiri — ayah (admin) yang membuatkan link untuk guru ngajinya |
| Siapa boleh melihat daftar & revoke | Pemilik dashboard + admin keluarga | Simetris dengan aturan pembuatan |
| Expiry default | **90 hari** (bisa pilih: 7 / 30 / 90 hari / tanpa batas) | Link yang terlupakan mati sendiri; "tanpa batas" tetap disediakan untuk kakek-nenek |
| Maksimum share aktif per anggota | **10** | Mencegah penumpukan; kebutuhan riil paling 2–3 |
| Perpanjangan | Tidak ada — buat link baru | Menjaga model sederhana; revoke + create = "perpanjang" |
| Anggota nonaktif (`users.is_active = false`) | Link-nya ikut mati (cek saat view) | Nonaktif berarti keluar dari pemantauan |

---

## 4. API

### 4.1 Endpoint terautentikasi (JWT)

```
POST   /api/v1/shares
       body: { "user_id": "…opsional, default diri sendiri…",
               "label": "Ustadz Ahmad",
               "expires_in_days": 90 }        ← 0/null = tanpa kedaluwarsa
       → 201 { "id", "label", "expires_at",
               "token": "…",                  ← HANYA muncul di sini, sekali
               "url": "https://…/s/…" }
       → 403 kalau user_id ≠ diri sendiri dan requester bukan admin
       → 400 kalau share aktif anggota tsb sudah 10

GET    /api/v1/shares                          ← daftar share milik saya
GET    /api/v1/shares?user_id=…                ← admin: daftar share anggota lain
       → 200 [ { id, user_id, user_name, label, expires_at,
                 last_viewed_at, view_count, status } ]
         status ∈ aktif | kedaluwarsa | dicabut  (dihitung, bukan kolom)
         (token TIDAK pernah ikut — sudah tidak bisa direkonstruksi dari hash)

DELETE /api/v1/shares/:id                      ← revoke (set revoked_at)
       → 200; 403 kalau bukan pemilik dashboard & bukan admin
       → idempotent: revoke dua kali tetap 200
```

### 4.2 Endpoint publik (tanpa JWT — satu-satunya di aplikasi selain /health)

```
GET /api/v1/public/shares/:token
    → 200 payload di bawah; 404 seragam untuk semua kegagalan (lihat §2.2)
    → side effect: update last_viewed_at, increment view_count
```

Payload publik — **didefinisikan eksplisit, bukan reuse struct internal**, supaya penambahan field internal di masa depan tidak otomatis bocor ke publik:

```json
{
  "success": true,
  "data": {
    "member_name": "Ahmad",
    "label": "Ustadz Ahmad",
    "stats": {
      "total_hafal_ayat": 342,
      "total_hafal_surah": 12,
      "percent_hafal": 78.5,
      "streak": 6
    },
    "murajaah_today": { "done": 2, "total": 3 },
    "surah_progress": [
      { "surah_number": 67, "surah_name": "Al-Mulk",
        "ayat_hafal": 30, "ayat_total": 30, "status": "hafal" }
    ],
    "last_activity": [
      { "date": "2026-07-19", "type": "sabak",
        "detail": "Al-Mulk ayat 21–30", "status": "hafal" }
    ],
    "generated_at": "2026-07-20T07:00:00Z"
  }
}
```

**Yang sengaja TIDAK ada di payload publik:** email, `user_id`, `last_seen_at`, data anggota lain, invite code keluarga, nama keluarga. `member_name` cukup nama panggilan (field `name` existing).

`last_activity`: 7 entri terakhir dari `hafalan_progress` (order `noted_at` desc). `surah_progress`: agregasi per surah dari hafalan ber-status `sedang`/`hafal`; nama surah + jumlah ayat diambil dari cache Quran API yang sudah ada (`quran:surah:{number}`).

### 4.3 Caching payload publik

Redis, key `share:view:{token_hash}`, TTL **5 menit**. Alasan: guru bisa refresh berkali-kali; tanpa cache tiap view memicu query hafalan + murajaah + panggilan metadata surah. Invalidasi tidak perlu presisi — data hafalan bukan real-time critical, staleness 5 menit bisa diterima. Revoke tidak menunggu TTL: cek validitas share tetap ke DB tiap request (query ringan by unique index), yang di-cache hanya **payload**-nya.

---

## 5. Struktur Implementasi Backend

Mengikuti pola Clean Architecture existing:

```
internal/domain/entity/dashboard_share.go          (baru)
internal/domain/repository/share_repository.go     (baru — interface)
    Create, FindByTokenHash, ListByUserID, CountActiveByUserID,
    Revoke(id), TouchView(id)
internal/repository/db/share_repository.go         (baru — GORM impl)
internal/usecase/share_usecase.go                  (baru)
    CreateShare(ctx, requesterID, requesterRole, input)
    ListShares(ctx, requesterID, requesterRole, targetUserID)
    RevokeShare(ctx, requesterID, requesterRole, shareID)
    ViewShared(ctx, token) → PublicDashboard        ← logic validasi + kompose payload
internal/delivery/http/handler/share_handler.go    (baru)
internal/delivery/http/route/share_route.go        (baru)
migrations/005_create_dashboard_shares.sql         (baru)
```

- `ViewShared` memakai ulang `HafalanRepository.FindByUserID` dan `MurajaahRepository.FindTodayByUserID` yang sudah ada; perhitungan stats mengikuti logika `GetMyDashboard` di `dashboard_usecase.go` — **ekstrak kalkulasi stats jadi fungsi bersama** supaya angka di dashboard pribadi dan di halaman share tidak pernah beda.
- Route publik didaftarkan **di luar** group JWT di `router.go`, dengan limiter middleware sendiri.
- Wiring manual di `cmd/api/main.go` + `router.go` (aturan Wapgo).

---

## 6. Frontend (Nuxt 3)

### 6.1 Halaman publik `/s/[token]`

- **Tanpa layout ber-auth** — layout `public` sendiri: header logo + nama anggota + label, tanpa navigasi aplikasi.
- Isi: kartu stats (ayat hafal, surah, persen, streak), progress muraja'ah hari ini, daftar progress per surah, aktivitas terakhir. Komponen visual bisa memakai ulang komponen kartu dashboard existing (props-based, tanpa store auth).
- Error state 404: halaman ramah "Link tidak ditemukan atau sudah tidak berlaku" — tanpa detail.
- `useHead`: `noindex`, title "Progress Hafalan — {nama}".
- Mobile-first (guru membukanya dari WhatsApp di HP).

### 6.2 Manajemen share

- **Halaman profil anggota** (diri sendiri) + **halaman admin per-anggota**: section "Link Pantau" — daftar share (label, status, terakhir dilihat, view count), tombol buat & cabut.
- Modal buat share: input label (wajib), pilihan masa berlaku (7/30/90 hari/tanpa batas).
- Setelah dibuat: tampilkan URL **sekali** dengan tombol "Salin" + tombol share WhatsApp (`https://wa.me/?text=…`) + peringatan "Simpan link ini — tidak bisa ditampilkan lagi".
- Konfirmasi sebelum revoke: "Guru/keluarga yang memegang link ini tidak akan bisa melihat progress lagi."

---

## 7. Testing

- **Unit (share_usecase):**
  - Create: sukses diri sendiri; admin untuk anggota lain; member untuk orang lain → ditolak; melebihi 10 aktif → ditolak; token hasil generate ≠ tersimpan (yang tersimpan hash-nya).
  - View: token valid; token salah / revoked / expired / user nonaktif → error identik (404 semantics); payload tidak memuat email/user_id (assert eksplisit).
  - Revoke: pemilik; admin; orang lain → 403; idempotent.
- **Integration** (tanpa mock DB): register → login → catat hafalan → buat share → buka endpoint publik tanpa auth → verifikasi stats cocok dengan `/dashboard/me` → revoke → endpoint publik jadi 404.
- **Manual:** buka link di browser tanpa login (incognito), Safari iOS + Chrome Android, share via WhatsApp.
- Coverage tetap ≥ 80%.

---

## 8. Butuh Keputusan Reviewer

1. **Expiry default 90 hari** dengan opsi tanpa batas — setuju? (Alternatif: wajib expiry, maksimal 1 tahun.)
2. **`last_activity` di payload publik** (7 catatan hafalan terakhir) — perlu, atau stats agregat saja? Ini data paling "detail" yang dibagikan.
3. **Tombol share WhatsApp** di modal — dipakai atau cukup tombol salin?
4. **Scope kebocoran yang diterima:** pemegang link melihat nama + seluruh progress hafalan anggota tsb. Kalau dirasa terlalu terbuka, bisa ditambah opsi "sembunyikan detail per-surah" per share — tapi usul saya jangan dulu (YAGNI).

## 9. Urutan Implementasi (setelah disetujui)

1. Migration `005` + entity + repository (interface & GORM).
2. Ekstrak kalkulasi stats dari `dashboard_usecase.go` jadi fungsi bersama.
3. `share_usecase.go` + unit test.
4. Handler + route (publik dengan limiter, privat dengan JWT) + wiring `main.go`/`router.go`.
5. Integration test.
6. Frontend: halaman publik `/s/[token]`, section manajemen di profil & admin.
7. Manual test mobile + incognito.

Estimasi: backend ±1 hari, frontend ±1 hari, testing ±setengah hari. Tidak menyentuh skema tabel existing sama sekali — risiko regresi terhadap fitur v1 mendekati nol.
