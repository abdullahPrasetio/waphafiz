# Rancangan: Multi-Group Membership (Keluarga + Sekolah)

> **Status:** DRAFT — menunggu review
> **Tanggal:** 2026-07-18
> **Scope:** Perubahan struktural di luar scope v1. Dokumen ini adalah rancangan v2, belum untuk dikerjakan sebelum disetujui.

---

## 1. Latar Belakang & Tujuan

### Kondisi sekarang (v1)

WAPHafiz dirancang untuk **satu keluarga** (10–20 orang):

- Satu user terikat ke **tepat satu** `family_group` lewat kolom `users.family_group_id` (nullable FK).
- Role (`admin` / `member`) melekat di **level user**, berlaku global.
- Invite code melekat 1:1 di `family_groups.invite_code`, dipakai sekali saat register.
- Semua data hafalan & muraja'ah di-scope per `user_id` — tidak menyentuh grup sama sekali.

### Kebutuhan baru

1. Konsep grup diperluas: bukan hanya **keluarga**, tapi juga **sekolah islam** (pesantren/madrasah/TPQ).
2. Yang mengelola grup bisa **admin keluarga** (kepala keluarga) atau **guru** (di sekolah).
3. **Satu user bisa terdaftar di lebih dari satu grup sekaligus** — contoh: seorang anak tercatat di grup keluarganya DAN di grup sekolahnya. Orang tuanya memantau lewat grup keluarga; gurunya memantau lewat grup sekolah.

### Prinsip desain kunci

> **Hafalan itu milik orang, bukan milik grup.**

Progress hafalan seorang anak adalah fakta tentang dirinya — kalau dia hafal Al-Mulk, dia hafal Al-Mulk baik dilihat dari sisi keluarga maupun sekolah. Karena itu:

- `hafalan_progress` dan `murajaah_schedule` **tetap di-scope per `user_id`**, TIDAK dipindah ke per-grup.
- Yang berubah menjadi per-grup hanyalah: **keanggotaan, role, dan visibilitas dashboard** (siapa boleh melihat progress siapa).

Konsekuensi enak dari prinsip ini: perubahan skema terkonsentrasi di area auth/membership saja. Business logic hafalan & muraja'ah (bagian paling kritis) **hampir tidak tersentuh**.

---

## 2. Perubahan Model Data

### 2.1 Ringkasan perubahan

| Tabel | Nasib |
|---|---|
| `family_groups` | **Rename konseptual → `groups`**, tambah kolom `type` (`family` / `school`) dan `description` |
| `users` | **Hapus** `family_group_id` dan `role` (pindah ke pivot). `is_active` tetap di user (lihat §2.5) |
| `group_memberships` | **Baru** — tabel pivot user ↔ group, menyimpan role per grup |
| `hafalan_progress` | Tidak berubah |
| `murajaah_schedule` | Tidak berubah |
| `murajaah_log` | Tidak berubah |

### 2.2 Tabel `groups` (generalisasi `family_groups`)

```sql
CREATE TABLE groups (
    id          UUID PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    type        VARCHAR(20)  NOT NULL DEFAULT 'family',  -- 'family' | 'school'
    description VARCHAR(500),                            -- opsional, mis. nama lengkap sekolah
    invite_code VARCHAR(50)  NOT NULL UNIQUE,
    created_by  UUID         NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_groups_type ON groups(type);
```

Entity Go:

```go
type GroupType string

const (
    GroupTypeFamily GroupType = "family"
    GroupTypeSchool GroupType = "school"
)

type Group struct {
    ID          uuid.UUID      `gorm:"type:uuid;primaryKey;not null" json:"id"`
    Name        string         `gorm:"size:200;not null"             json:"name"`
    Type        GroupType      `gorm:"size:20;not null;default:'family'" json:"type"`
    Description string         `gorm:"size:500"                      json:"description"`
    InviteCode  string         `gorm:"size:50;uniqueIndex;not null"  json:"invite_code"`
    CreatedBy   uuid.UUID      `gorm:"type:uuid;not null"            json:"created_by"`
    CreatedAt   time.Time      `                                     json:"created_at"`
    UpdatedAt   time.Time      `                                     json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index"                         json:"-"`
}

func (Group) TableName() string { return "groups" }
```

**Keputusan penamaan:** nama tabel fisik diganti `family_groups` → `groups` lewat migration `ALTER TABLE ... RENAME`. Alternatifnya adalah mempertahankan nama `family_groups` dan hanya menambah kolom `type` — lebih sedikit perubahan, tapi nama tabel jadi menyesatkan permanen ("family_groups berisi sekolah"). Karena aplikasi belum production dan datanya masih kecil, **rename sekarang lebih murah daripada menanggung nama salah selamanya**. → *butuh keputusan reviewer, lihat §9*.

### 2.3 Tabel `group_memberships` (baru — inti perubahan)

```sql
CREATE TABLE group_memberships (
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users(id),
    group_id   UUID        NOT NULL REFERENCES groups(id),
    role       VARCHAR(20) NOT NULL DEFAULT 'member',   -- 'admin' | 'member'
    is_active  BOOLEAN     NOT NULL DEFAULT true,        -- aktif di grup INI
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT uq_membership UNIQUE (user_id, group_id)
);

CREATE INDEX idx_memberships_user  ON group_memberships(user_id);
CREATE INDEX idx_memberships_group ON group_memberships(group_id);
```

Entity Go:

```go
type MembershipRole string

const (
    MembershipRoleAdmin  MembershipRole = "admin"
    MembershipRoleMember MembershipRole = "member"
)

type GroupMembership struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey;not null"          json:"id"`
    UserID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:uq_membership" json:"user_id"`
    GroupID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:uq_membership" json:"group_id"`
    Role      MembershipRole `gorm:"size:20;not null;default:'member'"      json:"role"`
    IsActive  bool           `gorm:"not null;default:true"                  json:"is_active"`
    JoinedAt  time.Time      `gorm:"not null"                               json:"joined_at"`
    CreatedAt time.Time      `                                              json:"created_at"`
    UpdatedAt time.Time      `                                              json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index"                                  json:"-"`

    User  *User  `gorm:"foreignKey:UserID"  json:"user,omitempty"`
    Group *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

func (GroupMembership) TableName() string { return "group_memberships" }
```

**Constraint penting:**

- `UNIQUE(user_id, group_id)` — satu user tidak bisa dobel-join grup yang sama. Catatan implementasi: karena pakai soft delete (GORM `deleted_at`), unique constraint polos akan menghalangi re-join setelah leave. Solusi: **partial unique index** `CREATE UNIQUE INDEX uq_membership ON group_memberships(user_id, group_id) WHERE deleted_at IS NULL;` (PostgreSQL mendukung ini).
- Satu user boleh punya banyak baris membership → **inilah yang mengaktifkan "1 anggota di keluarga DAN sekolah"**.
- Role berbeda per grup: user yang sama bisa `admin` di grup keluarganya tapi `member` di grup sekolah (atau sebaliknya — seorang guru adalah `admin` di grup sekolah dan `member` biasa di grup keluarganya sendiri).

### 2.4 Keputusan role: "guru" BUKAN role baru

Dua opsi dipertimbangkan:

| | Opsi A — role tetap `admin`/`member` | Opsi B — tambah role `guru` |
|---|---|---|
| Konsep | "Guru" = **label UI** untuk admin di grup ber-`type=school`. "Kepala keluarga" = label UI untuk admin di grup `family`. | Enum role jadi `admin` / `guru` / `member`, permission `guru` ≈ `admin` |
| Perubahan RBAC | Minimal — middleware cukup cek `admin` | Semua cek role harus di-update jadi `admin OR guru` |
| Fleksibilitas | Cukup, selama hak guru = hak admin keluarga | Perlu hanya jika kelak hak guru ≠ hak admin (mis. guru tidak boleh regenerate invite code) |
| Risiko | — | Enum tiga nilai yang dua di antaranya identik = sumber bug klasik (lupa cek salah satu) |

**Rekomendasi: Opsi A.** Role di database tetap dua (`admin`/`member`); presentasinya yang berubah:

```
group.type = 'family' → admin ditampilkan sebagai "Kepala Keluarga", member sebagai "Anggota"
group.type = 'school' → admin ditampilkan sebagai "Guru / Ustadz", member sebagai "Santri"
```

Mapping label ini murni urusan frontend (i18n/computed berdasar `group.type`). Kalau kelak kebutuhan hak akses guru benar-benar menyimpang dari admin keluarga, saat itulah role ketiga ditambahkan — jangan sekarang (YAGNI).

### 2.5 Perubahan tabel `users`

```sql
ALTER TABLE users DROP COLUMN family_group_id;
ALTER TABLE users DROP COLUMN role;
-- is_active DIPERTAHANKAN di users
```

- `family_group_id` → digantikan sepenuhnya oleh `group_memberships`.
- `role` → pindah ke `group_memberships.role` (role kini per-grup, bukan global).
- `is_active` → **ada di dua level dengan makna berbeda**:
  - `users.is_active` = akun dinonaktifkan global (tidak bisa login sama sekali). Dipertahankan.
  - `group_memberships.is_active` = dinonaktifkan **di grup itu saja** (mis. santri lulus/keluar dari sekolah, tapi akunnya tetap hidup dan grup keluarganya tetap jalan). Ini penting: guru menonaktifkan santri **tidak boleh** ikut mematikan akses anak itu di grup keluarganya.

Struct `User` setelah perubahan:

```go
type User struct {
    ID         uuid.UUID      `gorm:"type:uuid;primaryKey;not null"     json:"id"`
    Name       string         `gorm:"size:100;not null"                 json:"name"`
    Email      string         `gorm:"size:255;uniqueIndex;not null"     json:"email"`
    Password   string         `gorm:"size:255;not null"                 json:"-"`
    IsActive   bool           `gorm:"not null;default:true"             json:"is_active"`
    LastSeenAt *time.Time     `                                         json:"last_seen_at"`
    CreatedAt  time.Time      `                                         json:"created_at"`
    UpdatedAt  time.Time      `                                         json:"updated_at"`
    DeletedAt  gorm.DeletedAt `gorm:"index"                             json:"-"`

    Memberships []GroupMembership `gorm:"foreignKey:UserID" json:"memberships,omitempty"`
}
```

### 2.6 ERD sesudah perubahan

```
users ──1:N── group_memberships ──N:1── groups
  │                                    (type: family|school)
  ├──1:N── hafalan_progress        ← TIDAK BERUBAH, tetap per-user
  ├──1:N── murajaah_schedule       ← TIDAK BERUBAH, tetap per-user
  └──1:N── murajaah_log            ← TIDAK BERUBAH, tetap per-user
```

---

## 3. Konteks Grup di API (Group Scoping)

### 3.1 Masalah

Di v1, "grup aktif" implisit — user cuma punya satu. Setelah multi-group, setiap request yang menyentuh data grup harus tahu **grup mana yang dimaksud**. Ada tiga pendekatan umum:

| Pendekatan | Cara | Penilaian |
|---|---|---|
| Claim di JWT (`active_group_id`) | Grup aktif di-embed saat login/switch | ❌ Ganti grup = terbit token baru; token basi saat role berubah; menyulitkan buka 2 tab beda grup |
| Header `X-Group-ID` | Frontend kirim header di tiap request | ⚠️ Bekerja, tapi konteks jadi "tersembunyi", sulit di-debug/cache, tidak RESTful |
| **Path parameter `/groups/:groupId/...`** | Grup eksplisit di URL | ✅ Eksplisit, RESTful, bisa bookmark/share, middleware gampang validasi |

**Rekomendasi: path parameter**, dengan pengecualian: endpoint yang datanya per-user murni (hafalan diri sendiri, muraja'ah diri sendiri, quran proxy) **tidak butuh konteks grup sama sekali** dan URL-nya tidak berubah.

### 3.2 Aturan pembagian endpoint

```
┌──────────────────────────────────────┬──────────────────────────────┐
│ Data per-USER (grup tidak relevan)   │ Data per-GRUP                │
├──────────────────────────────────────┼──────────────────────────────┤
│ POST/GET/PATCH /hafalan              │ Daftar anggota grup          │
│ GET /hafalan/summary                 │ Progress semua anggota grup  │
│ GET /murajaah/today, /history        │ Kelola membership & role     │
│ POST /murajaah/:id/complete          │ Invite code                  │
│ GET /quran/**                        │ Dashboard admin/guru         │
│ GET /dashboard/me                    │                              │
└──────────────────────────────────────┴──────────────────────────────┘
```

Kolom kiri: **URL tidak berubah dari v1** — nol breaking change untuk alur harian anggota.
Kolom kanan: pindah ke bawah prefix `/groups/:groupId/`.

### 3.3 Daftar endpoint sesudah perubahan

```
# Auth — berubah perilaku, URL tetap
POST /api/v1/auth/register        body: {invite_code, name, email, password}
                                  → buat user + membership ke grup pemilik invite_code
POST /api/v1/auth/login
POST /api/v1/auth/logout

# Groups — BARU
GET    /api/v1/groups                          ← daftar grup milik user login (+ role di masing²)
POST   /api/v1/groups                          ← buat grup baru (family/school), creator jadi admin
POST   /api/v1/groups/join                     body: {invite_code}
                                               ← user login join grup tambahan (INI fitur kuncinya)
GET    /api/v1/groups/:groupId                 ← detail grup (member only)
DELETE /api/v1/groups/:groupId/membership      ← keluar dari grup (self-leave)

# Group management — pengganti /family/** (admin/guru grup tsb)
GET    /api/v1/groups/:groupId/members
PATCH  /api/v1/groups/:groupId/members/:userId/status   ← aktif/nonaktif DI GRUP INI
PATCH  /api/v1/groups/:groupId/members/:userId/role     ← promote/demote admin
DELETE /api/v1/groups/:groupId/members/:userId          ← keluarkan anggota
POST   /api/v1/groups/:groupId/invite-code/regenerate

# Dashboard & monitoring per grup — pengganti /admin/** (admin/guru grup tsb)
GET    /api/v1/groups/:groupId/dashboard
GET    /api/v1/groups/:groupId/hafalan          ← progress hafalan seluruh anggota grup
GET    /api/v1/groups/:groupId/members/:userId/hafalan
GET    /api/v1/groups/:groupId/members/:userId/murajaah

# Per-user (TIDAK BERUBAH dari v1)
POST/GET       /api/v1/hafalan
PATCH          /api/v1/hafalan/:id
GET            /api/v1/hafalan/summary
GET            /api/v1/murajaah/today
POST           /api/v1/murajaah/:id/complete
GET            /api/v1/murajaah/history
GET            /api/v1/quran/**
GET            /api/v1/dashboard/me
```

Endpoint v1 yang **dihapus** (diganti bentuk baru): `GET /family/members`, `PATCH /family/members/:id/status`, `POST /family/invite-code/regenerate`, `GET /admin/hafalan`, `GET /admin/dashboard`.

### 3.4 Contoh request/response kunci

**`GET /api/v1/groups`** — dipakai frontend untuk group switcher:

```json
{
  "success": true,
  "data": [
    {
      "id": "…",
      "name": "Keluarga Prasetio",
      "type": "family",
      "my_role": "admin",
      "member_count": 6,
      "joined_at": "2026-07-01T03:00:00Z"
    },
    {
      "id": "…",
      "name": "TPQ Al-Hikmah",
      "type": "school",
      "my_role": "member",
      "member_count": 34,
      "joined_at": "2026-07-15T08:30:00Z"
    }
  ]
}
```

**`POST /api/v1/groups/join`** dengan `{"invite_code": "ALHIKMAH24"}`:

- 200 → membership dibuat dengan role `member`, response berisi detail grup.
- 400 `"Kode undangan tidak valid"` → kode tidak ditemukan.
- 409 `"Kamu sudah terdaftar di grup ini"` → membership (aktif) sudah ada.
- Jika ada membership soft-deleted (pernah leave) → **restore** membership lama sebagai `member` (progress historis tak hilang karena memang tidak pernah bergantung pada membership).

---

## 4. Perubahan Auth & RBAC

### 4.1 JWT

Claim JWT **tidak lagi memuat role** (role kini per-grup, tidak ada "role global"). JWT cukup: `sub` (user_id), `exp`, dst. Ini menyederhanakan token dan menghindari token basi saat role diubah.

### 4.2 Middleware

`RequireRole("admin")` (global) **diganti** middleware baru berbasis membership:

```go
// RequireGroupMember: user login harus member aktif dari :groupId
// RequireGroupAdmin : user login harus member aktif dengan role admin di :groupId
//
// Keduanya:
// 1. Parse :groupId dari path
// 2. Query group_memberships WHERE user_id=? AND group_id=? AND is_active=true AND deleted_at IS NULL
// 3. Tidak ketemu           → 403 "Kamu bukan anggota grup ini"
// 4. (admin) role != admin  → 403 "Hanya admin/guru yang boleh mengakses ini"
// 5. Simpan membership di c.Locals("membership") agar handler tidak query ulang
```

Pemetaan proteksi:

| Route | Middleware |
|---|---|
| `GET /groups/:groupId` | `RequireGroupMember` |
| `GET /groups/:groupId/members` | `RequireGroupMember` (anggota boleh lihat sesama anggota) |
| `PATCH …/members/:userId/*`, `DELETE …/members/:userId`, `POST …/invite-code/regenerate` | `RequireGroupAdmin` |
| `GET /groups/:groupId/dashboard`, `…/hafalan`, `…/members/:userId/hafalan` | `RequireGroupAdmin` |
| Endpoint per-user (hafalan, murajaah, dashboard/me) | JWT saja, tanpa cek grup |

**Catatan visibilitas** *(butuh keputusan reviewer, §9)*: apakah `GET /groups/:groupId/members` + progress ringkas boleh dilihat semua anggota (di keluarga wajar: saling lihat progress = motivasi), atau khusus admin (di sekolah 30+ santri mungkin tidak semua santri perlu lihat progress santri lain)? Usulan default: **daftar member boleh semua anggota; detail hafalan per anggota hanya admin.** Bisa dibuat per-`group.type` kalau dirasa perlu, tapi mulai dari aturan tunggal dulu.

### 4.3 Register & aturan grup

- Register **tetap wajib invite code** (tidak ada open registration — konsisten dengan v1).
- Register = buat `users` + satu `group_memberships` (role `member`) ke grup pemilik kode, dalam **satu transaksi DB**.
- Pembuatan grup baru (`POST /groups`): tersedia untuk **semua user login** — use case nyata: seorang ayah yang tadinya cuma `member` di grup sekolah anaknya ingin membuat grup keluarganya sendiri; atau guru membuat grup sekolah. Creator otomatis jadi `admin` grup tsb. Invite code di-generate server (bukan input user).
- Admin terakhir tidak boleh leave/demote diri sendiri kalau grup masih punya anggota lain → 400 `"Angkat admin lain dulu sebelum keluar"`.

---

## 5. Dampak ke Business Logic Existing

### 5.1 Hafalan & Muraja'ah — **nol perubahan logika**

Seluruh usecase hafalan (overlap check, validasi ayat, status) dan muraja'ah (sabqi 7 hari, manzil rotasi mod 7, lazy generation, skip kalau kosong) beroperasi murni pada `user_id`. Tidak ada satupun yang membaca `family_group_id`. **Tidak disentuh.**

Ini juga berarti: anak yang tercatat di keluarga + sekolah punya **satu** progress hafalan dan **satu** jadwal muraja'ah — bukan dua. Guru dan orang tua melihat data yang sama dari jendela masing-masing. Ini perilaku yang diinginkan (bukan bug): hafalan tidak dobel dicatat.

### 5.2 Usecase yang berubah

| File | Perubahan |
|---|---|
| `auth_usecase.go` | Register: buat membership, bukan set `family_group_id`. Login: hapus role dari klaim/respons global, tambahkan daftar membership di respons login |
| `family_usecase.go` | **Rename → `group_usecase.go`**, semua operasi menerima `groupID` + validasi via membership; tambah: `ListMyGroups`, `CreateGroup`, `JoinByInviteCode`, `LeaveGroup`, `ChangeMemberRole` |
| `dashboard_usecase.go` | `AdminDashboard` → `GroupDashboard(groupID)`: agregasi progress di-JOIN lewat `group_memberships`, bukan `users.family_group_id` |
| `hafalan_usecase.go` | Hanya bagian admin listing (`GET /admin/hafalan` → `GET /groups/:groupId/hafalan`): filter berubah dari "semua user se-family" jadi "semua member grup ini" |
| `user_usecase.go` | Hapus asumsi role global; status aktif per-grup pindah ke group usecase |

### 5.3 Repository

- `family_group_repository.go` → `group_repository.go` (+ method by-type bila perlu).
- **Baru**: `membership_repository.go` — `FindByUserAndGroup`, `ListByUser`, `ListByGroup`, `Create`, `Restore`, `UpdateRole`, `UpdateStatus`, `SoftDelete`.
- `user_repository.go` — hapus query berbasis `family_group_id`.

Query kunci dashboard grup (pengganti filter `family_group_id`):

```sql
SELECT u.id, u.name, /* agregat hafalan */
FROM users u
JOIN group_memberships gm ON gm.user_id = u.id
    AND gm.group_id = $1
    AND gm.is_active = true
    AND gm.deleted_at IS NULL
LEFT JOIN hafalan_progress hp ON hp.user_id = u.id AND hp.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.name;
```

---

## 6. Rencana Migration

Empat file migration baru (lanjutan `004_…`), urutan penting:

```
005_create_groups_and_memberships.sql
    -- a. ALTER TABLE family_groups RENAME TO groups;
    -- b. ALTER TABLE groups ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'family';
    --    ALTER TABLE groups ADD COLUMN description VARCHAR(500);
    -- c. CREATE TABLE group_memberships (…);
    --    CREATE UNIQUE INDEX uq_membership ON group_memberships(user_id, group_id)
    --        WHERE deleted_at IS NULL;

006_backfill_memberships.sql
    -- Salin keanggotaan lama ke pivot:
    INSERT INTO group_memberships (id, user_id, group_id, role, is_active, joined_at, created_at, updated_at)
    SELECT gen_random_uuid(), u.id, u.family_group_id, u.role, u.is_active, u.created_at, now(), now()
    FROM users u
    WHERE u.family_group_id IS NOT NULL
      AND u.deleted_at IS NULL;

007_drop_users_group_columns.sql
    -- Setelah backfill terverifikasi:
    ALTER TABLE users DROP COLUMN family_group_id;
    ALTER TABLE users DROP COLUMN role;
```

Catatan:

- `006` mempertahankan `role` lama apa adanya: admin keluarga v1 → admin grup family. Tidak ada perubahan efektif hak akses bagi user existing.
- User v1 dengan `family_group_id IS NULL` (kalau ada) → jadi user tanpa membership; ditangani oleh UX §7.3.
- Migration `007` dipisah dari `006` supaya bisa verifikasi manual hasil backfill dulu (`SELECT count(*)` cocok) sebelum drop kolom. Rollback plan: selama `007` belum jalan, kolom lama masih utuh.
- Deploy backend baru dan migration harus satu paket (kode baru tidak membaca kolom lama; kode lama tidak tahu tabel baru) — karena user masih internal keluarga sendiri, downtime beberapa menit saat cutover bisa diterima; tidak perlu strategi dual-write.

---

## 7. Dampak Frontend (Nuxt 3)

### 7.1 Group switcher

Komponen baru di header/sidebar (desktop) dan di halaman profil (mobile):

```
┌────────────────────────┐
│ 🏠 Keluarga Prasetio ▾ │   ← grup aktif tersimpan di Pinia + localStorage
├────────────────────────┤
│ ✓ 🏠 Keluarga Prasetio │
│   🕌 TPQ Al-Hikmah     │
│ ──────────────────────│
│ ＋ Gabung grup lain    │   ← modal input invite code → POST /groups/join
│ ＋ Buat grup baru      │
└────────────────────────┘
```

- Grup aktif = konteks untuk halaman anggota & dashboard admin. `groupId` aktif dipakai menyusun URL API `/groups/:groupId/...`.
- Halaman hafalan & muraja'ah pribadi **tidak terpengaruh** grup aktif (datanya per-user).
- User dengan tepat satu grup: switcher tampil sebagai label statis (tanpa dropdown pilih, tapi tetap ada aksi "Gabung grup lain").

### 7.2 Label role kontekstual (lihat §2.4)

```ts
// composables/useGroupLabels.ts
const roleLabel = (role: string, groupType: string) =>
  groupType === 'school'
    ? role === 'admin' ? 'Guru' : 'Santri'
    : role === 'admin' ? 'Kepala Keluarga' : 'Anggota'
```

Ikon grup: `family` → `IconHome`, `school` → `IconBuildingMosque` (Tabler Icons, konsisten design system).

### 7.3 State kosong

- User tanpa membership sama sekali (edge case pasca-migration atau setelah leave semua grup): halaman onboarding "Gabung grup dengan kode undangan, atau buat grup baru". Fitur pribadi (hafalan, muraja'ah, quran) **tetap bisa dipakai** — sekali lagi, hafalan milik orang, bukan grup.

### 7.4 Store Pinia

- `stores/group.ts` (baru): `myGroups`, `activeGroupId` (persist localStorage), `activeMembership` (termasuk `my_role` — menentukan menu admin tampil/tidak).
- `stores/auth.ts`: hapus `role` global; menu admin kini derived dari `activeMembership.role`.

---

## 8. Testing

- **Unit (usecase)**: join by invite code (sukses / kode salah / sudah member / re-join setelah leave), leave grup (member biasa / admin terakhir ditolak), change role, guard `RequireGroupAdmin` lintas grup (admin grup A akses grup B → 403), dashboard grup hanya menghitung member aktif grup tsb.
- **Integration** (sesuai aturan: tanpa mock DB): alur penuh — register via invite keluarga → login → catat hafalan → buat grup sekolah → user kedua join sekolah → user pertama join sekolah juga → verifikasi: (a) muncul di dashboard kedua grup, (b) progress hafalannya identik dilihat dari kedua grup, (c) nonaktif di sekolah tidak mempengaruhi akses grup keluarga.
- **Migration test**: jalankan `005`–`007` terhadap snapshot data v1, verifikasi jumlah membership = jumlah user ber-`family_group_id`, role terbawa benar.
- Coverage backend tetap ≥ 80% (`make coverage`).
- Manual mobile: group switcher di Safari iOS & Chrome Android.

---

## 9. Butuh Keputusan Reviewer

1. **Rename tabel `family_groups` → `groups`** (rekomendasi: ya, selagi murah) — atau pertahankan nama lama + kolom `type`?
2. **Role "guru" sebagai label UI dari `admin`** (rekomendasi, §2.4) — atau benar-benar role ketiga di enum?
3. **Visibilitas antar-anggota** (§4.2): daftar member terlihat semua anggota + detail hafalan hanya admin — setuju? Perlu dibedakan per tipe grup?
4. **Siapa boleh buat grup**: semua user login (rekomendasi, §4.3) — atau dibatasi?
5. **Update scope**: dokumen ini keluar dari "Scope v1 — Frozen" di CLAUDE.md. Kalau disetujui, CLAUDE.md (root + backend) dan prd.md perlu di-update bersamaan dengan implementasi — konfirmasi ini jadi bagian dari pekerjaan.

## 10. Urutan Implementasi (setelah disetujui)

1. Migration `005`–`007` + entity `Group`, `GroupMembership`, perubahan `User`.
2. `membership_repository` + rename `group_repository`; sesuaikan `user_repository`.
3. Middleware `RequireGroupMember` / `RequireGroupAdmin`; bersihkan role dari JWT.
4. Usecase: auth (register/login), group (join/create/leave/role), dashboard, admin-hafalan-listing.
5. Routes & handlers baru; hapus route `/family/**`, `/admin/**`.
6. Unit + integration test; migration test.
7. Frontend: store group, group switcher, label kontekstual, halaman join/create, onboarding kosong.
8. Update dokumen: CLAUDE.md (root & backend), prd.md, devplan.md.

Estimasi ukuran: perubahan backend terkonsentrasi di auth/group/dashboard (~5 usecase, 2 repo, 3 migration); frontend 1 store + 1 komponen switcher + penyesuaian halaman admin. Logika inti hafalan/muraja'ah tidak tersentuh.
