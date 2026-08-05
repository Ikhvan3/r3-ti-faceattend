# R3 TI FaceAttend

R3 TI FaceAttend adalah prototype absensi lokal untuk pegawai Divisi Teknologi
Informasi PTPN I Regional 3 Semarang.

## Struktur Repository

- `mobile/` aplikasi Flutter Android untuk pegawai.
- `admin-web/` website admin berbasis Next.js App Router.
- `backend/` REST API utama berbasis Golang.
- `docs/` dokumentasi teknis proyek.

## Batas Tahap Ini

Prototype saat ini mencakup autentikasi admin, backend Employee Management API,
antarmuka admin Next.js untuk manajemen pegawai Divisi Teknologi Informasi, dan
autentikasi mobile untuk pegawai `USER` aktif. Modul absensi, geolocation,
jadwal, lokasi kerja, face recognition, liveness, laporan, dan dashboard
statistik belum masuk scope tahap ini.

## Komponen Lokal

- Flutter untuk aplikasi Android.
- Next.js dengan TypeScript, Tailwind CSS, ESLint, App Router, dan `src/`.
- Golang REST API tanpa framework HTTP dan tanpa ORM.
- PostgreSQL hanya diakses oleh backend Golang.

## Mobile

Aplikasi mobile memakai endpoint backend `/api/v1/auth/*` untuk login,
restore session, refresh token, dan logout. Base URL default untuk Android
emulator adalah `http://10.0.2.2:8080/api/v1` dan dapat diubah melalui
`--dart-define=API_BASE_URL=...`.

Mobile hanya menerima user dengan `role = USER` dan `account_status = ACTIVE`.
Token disimpan di secure storage, bukan di storage biasa atau log.

## Navigasi Admin Web

Setelah login admin, modul yang tersedia:

- `/dashboard`
- `/employees`
- `/employees/new`
- `/employees/{id}`
- `/employees/{id}/edit`

Gunakan hanya data dummy pada development. Jangan memakai data pegawai PTPN
yang sebenarnya.

## Dokumentasi

- [Arsitektur](docs/architecture.md)
- [Roadmap](docs/roadmap.md)
- [Development](docs/development.md)
