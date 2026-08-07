# R3 TI FaceAttend

R3 TI FaceAttend adalah prototype absensi lokal untuk pegawai Divisi Teknologi.

## Struktur Repository

- `mobile/` aplikasi Flutter Android untuk pegawai.
- `admin-web/` website admin berbasis Next.js App Router.
- `backend/` REST API utama berbasis Golang.
- `docs/` dokumentasi teknis proyek.

## Batas Tahap Ini

Prototype saat ini mencakup autentikasi admin, backend Employee Management API,
antarmuka admin Next.js untuk manajemen pegawai Divisi Teknologi Informasi,
manajemen jadwal kerja, penugasan jadwal, manajemen lokasi kantor, penugasan
lokasi, autentikasi mobile untuk pegawai `USER` aktif, absensi dasar, dan
enforcement geofence server-side untuk check-in/check-out mobile. Face
enrollment dan face verification standalone sudah tersedia. Active face
liveness masih berupa prototype standalone untuk uji keaktifan wajah dan belum
menjadi syarat attendance. Laporan dan dashboard statistik belum masuk scope
tahap ini.

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
- `/work-schedules`
- `/work-schedules/new`
- `/work-schedules/{id}`
- `/work-schedules/{id}/edit`
- `/schedule-assignments`
- `/schedule-assignments/new`
- `/schedule-assignments/{id}`
- `/office-locations`
- `/office-locations/new`
- `/office-locations/{id}`
- `/office-locations/{id}/edit`
- `/location-assignments`
- `/location-assignments/new`
- `/location-assignments/{id}`

Mutasi jadwal kerja, penugasan jadwal, lokasi kantor, dan penugasan lokasi
melewati Next.js BFF di `/api/admin/*`. Token tetap berada di HttpOnly cookie
dan semua data tetap melewati Golang REST API.

## Dokumentasi

- [Arsitektur](docs/architecture.md)
- [Roadmap](docs/roadmap.md)
- [Development](docs/development.md)
