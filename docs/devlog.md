# Devlog — WAPHafiz

Catatan perkembangan harian pengerjaan project.

---

## 2026-06-17

### Yang Dikerjakan

**Backend — Hafalan CRUD (Delete & Admin Endpoints)**
- Tambah method `Delete` ke interface `HafalanRepository` dan implementasi soft-delete di GORM (`UPDATE SET deleted_at = NOW()`)
- Extend `HafalanUseCase` interface dengan: `Delete`, `AdminListByMember`, `AdminUpdate`, `AdminDelete`
- Implementasi usecase:
  - `Delete` — cek ownership sebelum hapus, return `ErrNotOwner` kalau bukan miliknya
  - `AdminUpdate` / `AdminDelete` — tanpa cek ownership (admin privilege)
  - `AdminListByMember` — ambil semua hafalan milik member tertentu
- Tambah handler: `Delete`, `AdminListByMember`, `AdminUpdate`, `AdminDelete` di `hafalan_handler.go`
- Daftarkan route baru di `hafalan_route.go`:
  - `DELETE /hafalan/:id`
  - `GET /admin/hafalan/member/:userID`
  - `PATCH /admin/hafalan/:id`
  - `DELETE /admin/hafalan/:id`
- Fix CORS: tambah `PATCH` ke `AllowMethods` di `middleware/cors.go` (sebelumnya hanya GET/POST/PUT/DELETE/OPTIONS)
- Fix test: tambah method `Delete` ke `fakeHafalanRepo` di `hafalan_usecase_test.go` agar compile

**Frontend — Halaman Hafalan (`/hafalan`)**
- Ganti wrapper card dari `NuxtLink` ke `div` — sebelumnya klik card langsung navigasi ke Al-Quran
- Tambah ikon `IconExternalLink` kecil di pojok kanan atas card untuk navigasi ke surah
- Ganti dropdown status lebar dengan 3 tombol toggle compact (Hafal/Sedang/Belum)
- Tambah tombol hapus (trash icon) per card dengan konfirmasi
- Definisikan `STATUS_OPTIONS` di script (bukan inline template) agar tidak error Vue parser TypeScript cast
- Tambah state `updatingId` dan `deletingId` untuk loading per-item

**Frontend — Halaman Admin Progress (`/admin/progress`)**
- Tambah tombol "Kelola" per member card
- Modal kelola hafalan: fetch semua hafalan member via `GET /admin/hafalan/member/:userID`
- Dropdown status per item auto-save via `PATCH /admin/hafalan/:id`
- Tombol hapus per item via `DELETE /admin/hafalan/:id`
- `onStatusChange` di script (bukan template) untuk handle `HTMLSelectElement` cast

**Frontend — Dashboard (`/dashboard`)**
- Fix `NaN%` di "Progress Hafalan Saya" — field `surahName`/`totalAyat` tidak ada di `HafalanProgress`
- Group hafalan per surah (`hafalanBySurah` computed) sebelum hitung persentase — sebelumnya tiap record tampil sendiri
- Fetch `quranStore.fetchSurahList()` di `onMounted` untuk data total ayat per surah
- Fix field muraja'ah di template: `surah_number`, `ayat_start`, `ayat_end`, `completed_at` (snake_case sesuai store)

**Frontend — Nuxt Config**
- Tambah `components: [{ path: '~/components', pathPrefix: false }]` di `nuxt.config.ts`
- Fix: `AppBadge`, `AppProgressBar` tidak bisa di-resolve karena Nuxt 3 default prefix path (`ui/AppBadge` → `UiAppBadge`)

### Bug / Error yang Ditemukan dan Diperbaiki

| Error | Penyebab | Fix |
|---|---|---|
| CORS error pada DELETE/PATCH | `PATCH` tidak ada di `AllowMethods` | Tambah `PATCH` ke cors middleware |
| `NaN%` di dashboard | Akses `h.surahName`, `h.totalAyat` yang tidak ada di type | Pakai `quranStore` + group by surah |
| `AppBadge` not resolved | Nuxt path prefix naming convention | `pathPrefix: false` di nuxt.config.ts |
| Klik card → navigasi ke Al-Quran | Card wrapper pakai `NuxtLink` | Ganti ke `div`, pisah link ke ikon |
| `as const` error di Vue template | TypeScript cast tidak bisa di template | Pindah ke script sebagai const |
| `fakeHafalanRepo` tidak compile | Interface `Delete` baru belum ada di fake | Tambah method ke test fake |
