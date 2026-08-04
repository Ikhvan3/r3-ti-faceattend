# R3 TI FaceAttend

R3 TI FaceAttend adalah prototype absensi lokal untuk pegawai Divisi Teknologi
Informasi PTPN I Regional 3 Semarang.

## Struktur Repository

- `mobile/` aplikasi Flutter Android untuk pegawai.
- `admin-web/` website admin berbasis Next.js App Router.
- `backend/` REST API utama berbasis Golang.
- `docs/` dokumentasi teknis proyek.

## Batas Tahap Ini

Tahap ini hanya menyiapkan struktur awal repository. Belum ada login, database
schema, CRUD pegawai, absensi, geolocation, face recognition, dashboard admin,
atau fitur bisnis lain.

## Komponen Lokal

- Flutter untuk aplikasi Android.
- Next.js dengan TypeScript, Tailwind CSS, ESLint, App Router, dan `src/`.
- Golang module minimal tanpa framework HTTP dan tanpa dependency database.
- PostgreSQL akan digunakan pada tahap berikutnya melalui backend Golang.

## Dokumentasi

- [Arsitektur](docs/architecture.md)
- [Roadmap](docs/roadmap.md)
- [Development](docs/development.md)
