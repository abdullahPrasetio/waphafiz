# WAPHafiz Frontend — Rules for Claude

Stack: **Nuxt 3 · Vue 3 · Tailwind CSS · Pinia · @vueuse/nuxt**

Baca `design/design.md` di root project untuk spesifikasi warna, komponen, dan layout.
Baca `CLAUDE.md` di root untuk rules project-wide dan domain terminology.

---

## Struktur Direktori (Target)

```
frontend/
├── assets/css/         ← global styles, CSS variables
├── components/
│   ├── ui/             ← komponen generik (Button, Badge, Card, Avatar, Input)
│   ├── hafalan/        ← SurahCard, HafalanModal, ProgressBar
│   ├── murajaah/       ← MurajaahCard, SummaryChip
│   ├── quran/          ← AyahCard, AudioPlayer, SurahHeader
│   └── layout/         ← AppSidebar, BottomNav, AppHeader
├── composables/        ← useAuth, useHafalan, useMurajaah, useQuran
├── layouts/
│   ├── default.vue     ← sidebar + main content (desktop)
│   └── auth.vue        ← halaman login/register (centered, no sidebar)
├── pages/
│   ├── login.vue
│   ├── register.vue
│   ├── dashboard.vue   ← / redirect ke sini setelah login
│   ├── quran/
│   │   ├── index.vue   ← daftar 114 surah
│   │   └── [number].vue← detail surah + ayat
│   ├── hafalan/
│   │   └── index.vue
│   ├── murajaah/
│   │   ├── index.vue
│   │   └── history.vue
│   └── admin/
│       ├── members.vue
│       ├── progress.vue
│       └── dashboard.vue
├── stores/
│   ├── auth.ts         ← JWT, user info, role
│   ├── hafalan.ts
│   ├── murajaah.ts
│   └── quran.ts
└── utils/
    ├── api.ts          ← base fetch wrapper dengan auth header
    └── quran.ts        ← helper juz calculation, ayat formatting
```

---

## Rules Wajib

### Komponen

- Komponen UI generik (Button, Badge, Card) taruh di `components/ui/` — jangan duplikat inline style.
- Nama komponen: PascalCase. File: `KebabCase.vue` atau `PascalCase.vue` (pilih satu, konsisten).
- Jangan pakai `scoped` CSS di komponen yang menggunakan Tailwind — gunakan Tailwind class langsung.
- Teks Arab **wajib** pakai font Amiri dan `dir="rtl"`: `class="font-arabic text-right"`.

### State Management (Pinia)

- Satu store per domain: `auth`, `hafalan`, `murajaah`, `quran`.
- Store hanya boleh fetch dari API lewat composable atau langsung — jangan taruh business logic di store.
- JWT disimpan di store (memory) + `localStorage` untuk persist session. Jangan simpan di cookie tanpa `httpOnly`.
- Saat token expired (401 dari API): clear store auth → redirect ke `/login`. Jangan retry infinite loop.

### API Calls

- Semua request HTTP lewat `utils/api.ts` — wrapper `$fetch` dengan base URL dari `runtimeConfig` dan header Authorization.
- Jangan hard-code `http://localhost:8080` di komponen atau store — pakai `useRuntimeConfig().public.apiBase`.
- Handle loading state dan error state di setiap fetch. Tampilkan pesan error yang jelas ke user.
- Gunakan `useAsyncData` atau `useFetch` untuk fetch saat SSR; gunakan `$fetch` untuk aksi user (button click).

### Routing & Auth Guard

- Route guard global: cek `authStore.isLoggedIn`. Jika false, redirect ke `/login`.
- Halaman admin (prefix `/admin/`): cek `authStore.user.role === 'admin'`. Jika bukan admin, redirect ke `/dashboard`.
- Setelah login sukses: redirect ke `/dashboard`.
- Setelah logout: clear store + redirect ke `/login`.

### Tailwind & Responsive

Ikuti breakpoint ini secara konsisten:

| Elemen | Mobile (default) | Desktop (`md:`) |
|---|---|---|
| Sidebar | `hidden` | `flex` |
| Bottom nav | `flex` | `hidden` |
| Stats grid | `grid-cols-2` | `md:grid-cols-4` |
| Surah grid | `grid-cols-2` | `md:grid-cols-3 lg:grid-cols-4` |
| Summary chips | `grid-cols-2` | `md:grid-cols-4` |
| Main padding | `p-5` | `md:p-8` |
| Grid-2 (dashboard) | `grid-cols-1` | `md:grid-cols-2` |

Warna utama: gunakan class Tailwind custom atau CSS variable dari `assets/css/`. Jangan hard-code hex warna di template.

### CSS Variables (setup di `assets/css/main.css`)

```css
:root {
  --green: #1D9E75;
  --green-light: #E1F5EE;
  --green-dark: #0F6E56;
  /* ... lihat design/design.md untuk daftar lengkap */
}
```

Di Tailwind config, extend warna dari CSS variable:
```js
// tailwind.config.ts
colors: {
  green: { DEFAULT: 'var(--green)', light: 'var(--green-light)', dark: 'var(--green-dark)' }
}
```

### Form & Validasi

- Validasi di sisi client sebelum submit: field required, format email, password min 8 karakter, invite code format `FAM-YYYY-XXXX`.
- Tampilkan error inline di bawah field, bukan hanya alert.
- Disable tombol submit saat loading untuk cegah double submit.

---

## Komponen Kritis

### AudioPlayer

- Gunakan HTML5 `<audio>` element — jangan library besar.
- URL audio dari `GET /api/v1/quran/ayah/:surah/:ayat/audio` (backend sudah proxy + cache).
- State: `playing`, `currentAyah`, `currentTime`, `duration`.
- Di mobile: collapse ke mini-bar di bawah layar saat scroll.

### AyahCard

```vue
<!-- Props yang wajib ada -->
:number="ayah.number"
:arabic="ayah.text"
:translation="ayah.translation"
:status="ayah.hafalanStatus"   <!-- 'belum' | 'sedang' | 'hafal' | null -->
:isPlaying="currentAyah === ayah.number"
@play="playAyah(ayah)"
@bookmark="toggleBookmark(ayah)"
@markHafal="openHafalanModal(ayah)"
```

### HafalanModal

- Input: surah (select), ayat_start (number), ayat_end (number), status (select)
- Validasi: `ayat_start <= ayat_end` sebelum submit
- Submit ke `POST /api/v1/hafalan`

### SurahCard (Hafalan Tracker)

- Border color berdasarkan status: hafal → `border-green-mid`, sedang → `border-amber`, belum → default
- Tampilkan: nomor surah, nama ID, nama Arab (Amiri), progress bar, jumlah ayat

---

## Halaman — Behavior

### `/dashboard`

- Fetch `GET /api/v1/dashboard/me` saat mount
- Tampilkan reminder bar jika muraja'ah hari ini belum selesai semua
- Streak chip: tampil hanya jika streak > 0

### `/quran/[number]`

- Fetch detail surah dari `GET /api/v1/quran/surah/:number`
- Audio player bar sticky di atas list ayat
- Highlight ayah yang sedang diputar (border green)
- Tombol "Tandai Hafal" buka `HafalanModal`

### `/hafalan`

- Default tampil semua surah yang pernah diinput user
- Filter: Semua / Hafal / Sedang / Belum — filter client-side (data sudah di-fetch semua)
- Search: filter by nama surah — client-side

### `/murajaah`

- Fetch `GET /api/v1/murajaah/today` saat mount
- Jika belum ada hafalan sama sekali: tampilkan empty state dengan link ke `/hafalan`
- Tombol "Selesai" hit `POST /api/v1/murajaah/:id/complete` → update UI langsung (optimistic update)

### `/admin/members`

- Hanya bisa diakses role `admin`
- Invite code: tombol copy ke clipboard + tombol regenerate (dengan konfirmasi)
- Tombol aktif/nonaktif anggota dengan konfirmasi sebelum eksekusi

---

## Error & Loading State

- Setiap halaman: tampilkan skeleton loading saat fetch pertama
- Error dari API: tampilkan toast notification (bukan alert browser)
- Error 503 dari Quran API: tampilkan "Audio tidak tersedia saat ini" — jangan sembunyikan
- Empty state yang informatif: jangan tampilkan halaman kosong tanpa penjelasan

---

## Nuxt Config Penting

```ts
// nuxt.config.ts
runtimeConfig: {
  public: {
    apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1'
  }
}
```

```bash
# .env
NUXT_PUBLIC_API_BASE=http://localhost:8080/api/v1
```

---

## Commands

```bash
npm run dev        # development server (port 3000)
npm run build      # production build
npm run preview    # preview production build
npm run lint       # ESLint
nuxi add page <name>
nuxi add component <name>
```
