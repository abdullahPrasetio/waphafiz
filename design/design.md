# WAPHafiz — Design System & UI Reference

Dokumen ini merangkum design system dan spesifikasi UI dari `waphafiz-design_1.html`.
Gunakan sebagai referensi saat membangun komponen di Nuxt 3 + Tailwind.

---

## Color Palette

| Token | Value | Penggunaan |
|---|---|---|
| `--green` | `#1D9E75` | Primary action, active state, progress bar |
| `--green-light` | `#E1F5EE` | Background badge hijau, highlight row |
| `--green-mid` | `#5DCAA5` | Border card hafal, hover state |
| `--green-dark` | `#0F6E56` | Hover button primary, text di atas green-light |
| `--green-darker` | `#085041` | Text label badge hijau gelap |
| `--blue-light` | `#E6F1FB` | Background badge manzil |
| `--blue-mid` | `#185FA5` | Text badge manzil |
| `--purple-light` | `#EEEDFE` | Background badge admin/role |
| `--purple-mid` | `#3C3489` | Text badge admin/role |
| `--amber-light` | `#FAEEDA` | Background badge sedang, streak chip |
| `--amber-mid` | `#633806` | Text badge sedang, streak chip |
| `--gray-bg` | `#F7F8FA` | Background halaman, background tabel header |
| `--gray-surface` | `#FFFFFF` | Card surface, sidebar, modal |
| `--gray-border` | `#E5E7EB` | Border card, input, divider |
| `--gray-border-soft` | `#F0F2F5` | Divider antar list item |
| `--text-primary` | `#111827` | Heading, label utama |
| `--text-secondary` | `#6B7280` | Subtext, placeholder |
| `--text-tertiary` | `#9CA3AF` | Helper text, icon inactive |

---

## Typography

- **Font utama:** `Inter` (400, 500, 600)
- **Font Arab:** `Amiri` (400, 700) — class `arabic`, direction `rtl`
- **Body default:** 13–14px Inter
- **Heading halaman:** 20–22px, weight 600
- **Label seksi:** 10–11px, uppercase, letter-spacing 0.07em, warna `--text-tertiary`

---

## Border Radius

| Token | Value | Dipakai pada |
|---|---|---|
| `--radius-sm` | `6px` | Badge, select qori, input kecil |
| `--radius-md` | `10px` | Button, input, chip, modal kecil |
| `--radius-lg` | `14px` | Card, player bar, ayah card, tabel |
| `--radius-xl` | `20px` | Auth card, device frame mobile |

---

## Komponen

### Badge

```html
<span class="badge badge-green">Hafal</span>
<span class="badge badge-blue">Manzil</span>
<span class="badge badge-amber">Sedang</span>
<span class="badge badge-gray">Belum</span>
<span class="badge badge-purple">Admin</span>
```

- Font: 11px, weight 500
- Padding: `3px 9px`, border-radius `20px`

### Button

```html
<!-- Primary -->
<button class="btn-primary">Simpan</button>

<!-- Outline -->
<button class="btn-outline">Batal</button>
```

- Primary: bg `--green`, hover `--green-dark`, text white, radius `--radius-md`
- Outline: border `--gray-border`, text `--text-secondary`, transparent bg

### Card

```html
<div class="card">
  <div class="card-header">
    <span class="card-title">Judul</span>
    <a class="card-link">Lihat semua →</a>
  </div>
  <!-- content -->
</div>
```

- Border: 1px `--gray-border`, radius `--radius-lg`, padding `1.25rem`

### Avatar

```html
<div class="avatar av-green" style="width:32px;height:32px;font-size:12px">AP</div>
```

Varian warna: `av-green`, `av-purple`, `av-amber`, `av-blue`

### Form Input

```html
<div class="form-field">
  <label class="form-label">Label</label>
  <div class="input-wrap">
    <i class="ti ti-mail"></i>
    <input class="form-input" type="email" placeholder="...">
  </div>
</div>
```

- Icon kiri: posisi absolute, `left: 10px`
- Input padding: `9px 12px 9px 36px`
- Focus ring: `border-color: --green`, `box-shadow: 0 0 0 3px rgba(29,158,117,0.12)`

### Progress Bar

```html
<div class="prog-bar-bg" style="height:5px">
  <div class="prog-bar" style="height:5px;width:74%"></div>
</div>
```

- Sedang dihafal (amber): override `background:#EF9F27` di bar fill

---

## Layout

### Desktop

```
┌─────────────────────────────────────────────┐
│  Sidebar (224px) │  Main Content (flex: 1)  │
│  sticky          │  background: --gray-bg    │
│  height: 100vh   │  padding: 2rem            │
└─────────────────────────────────────────────┘
```

- `app-layout`: `display: flex; min-height: 100vh`
- Sidebar sticky: `top: 49px` (tinggi tab-nav di design preview)

### Sidebar Nav Item

```html
<div class="nav-item active">
  <i class="ti ti-layout-dashboard"></i> Dashboard
</div>
```

- Active: bg `--green-light`, color `--green`, `border-left: 2px solid --green`
- Icon size: 18px

### Sidebar User Footer

```html
<div class="sidebar-user">
  <div class="avatar av-green">AP</div>
  <div>
    <div class="sidebar-user-name">Abdullah</div>
    <div class="sidebar-user-role">Admin</div>
  </div>
</div>
```

---

## Halaman

### Login & Register

- Wrapper: `min-height: 100vh`, centered flex
- Card: `max-width: 400px`, radius `--radius-xl`, padding `2.5rem`
- Logo mark: 40×40px, bg `--green`, rounded `--radius-md`
- Register: tampilkan banner invite-info (bg `--green-light`) di atas form

### Dashboard

- Streak chip: bg `--amber-light`, color `--amber-mid`, dengan icon `ti-flame`
- Reminder bar: bg `--green-light`, border `#9FE1CB`, flex row dengan icon bell
- Stats grid: `grid-template-columns: repeat(4, 1fr)` → 2 kolom di ≤768px
- Grid-2: 2 kartu sejajar (muraja'ah hari ini + progress hafalan)

### Al-Quran

- Surah header: tampilkan nama ID, nama Arab (font Amiri 32px, warna `--green`), info surah, progress bar
- Audio player bar: flex row — tombol play bulat (38×38px, bg `--green`), info qori, timeline, select qori
- Ayah card playing: `border-color: --green-mid`
- Teks Arab per ayah: font Amiri 24px, `text-align: right`, `direction: rtl`, `line-height: 2`

### Hafalan Tracker

- Toolbar: search + filter button (Semua/Hafal/Sedang/Belum) + tombol Tambah
- Surah grid: `grid-template-columns: repeat(auto-fill, minmax(175px, 1fr))`
- Card status border: hafal → `--green-mid`, sedang → `#FAC775`, belum → default
- Modal tambah hafalan: backdrop rgba overlay, modal box `max-width: 380px`

### Muraja'ah

- Summary chips: grid 4 kolom, card putih kecil
- Seksi label: uppercase, 11px, tertiary
- `mj-card.done`: `opacity: 0.6`
- Icon sabqi: bg `--green-light`, warna `--green`
- Icon manzil: bg `--blue-light`, warna `--blue-mid`

### Admin — Manajemen Anggota

- Stats 4 kartu di atas
- Invite code chip: inline, dengan tombol copy + regenerate
- Tabel: header uppercase 11px, bg `--gray-bg`; row hover bg `--gray-bg`
- Progress cell: bar 72×4px + label teks
- `last-seen.old`: warna `--text-tertiary`

---

## Mobile (Responsive)

### Breakpoint Rules

| Breakpoint | Perubahan |
|---|---|
| `≤ 768px` | Sidebar hilang; stats 2 kolom; grid-2 jadi 1 kolom; player timeline wrap ke bawah |
| `≤ 480px` | Stats 1 kolom; surah grid 2 kolom; toolbar kolom vertikal |

### Tailwind Equivalent (Nuxt)

```
Sidebar:        hidden md:flex
Bottom nav:     flex md:hidden
Stats grid:     grid-cols-2 md:grid-cols-4
Surah grid:     grid-cols-2 md:grid-cols-3 lg:grid-cols-4
Main padding:   p-5 md:p-8
```

### Bottom Navigation (Mobile)

```html
<div class="m-bottom-nav">
  <div class="m-nav-tab active">
    <i class="ti ti-layout-dashboard"></i>
    Dashboard
  </div>
  <!-- Quran, Hafalan, Muraja'ah, Profil -->
</div>
```

- 5 tab: Dashboard, Quran, Hafalan, Muraja'ah, Profil
- Active: color `--green`
- Icon 20px, label 9px

---

## Icon Library

Menggunakan **Tabler Icons** (`@tabler/icons-webfont`).

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@tabler/icons-webfont@latest/tabler-icons.min.css">
```

Di Nuxt, install: `npm install @tabler/icons-vue` untuk komponen Vue.

Ikon yang dipakai:
- `ti-layout-dashboard`, `ti-book`, `ti-checklist`, `ti-refresh`
- `ti-users`, `ti-chart-bar`, `ti-settings`
- `ti-mail`, `ti-lock`, `ti-user`, `ti-key`
- `ti-player-play`, `ti-player-pause`, `ti-bookmark`
- `ti-check`, `ti-plus`, `ti-search`, `ti-flame`
- `ti-bell`, `ti-arrow-left`, `ti-book-2`, `ti-calendar-check`
- `ti-dots`, `ti-copy`

---

## Quran — Data & Audio

- API: `alquran.cloud`
- Audio per ayat: `GET https://api.alquran.cloud/v1/ayah/{ref}/ar.alafasy`
- Pilihan qori default: Alafasy, Abdul Basit, Al-Husary
- Teks Arab dari API, terjemahan opsional (tampilkan di bawah teks Arab)
