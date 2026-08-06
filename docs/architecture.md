# Arsitektur

R3 TI FaceAttend menggunakan arsitektur lokal dengan tiga aplikasi utama:

- Flutter mobile untuk pegawai TI.
- Next.js admin website untuk administrator.
- Golang REST API sebagai satu-satunya pintu akses data.

PostgreSQL hanya boleh diakses oleh backend Golang. Flutter dan Next.js tidak
boleh melakukan koneksi langsung ke database.

## Batas Sistem Saat Ini

- Satu divisi: Teknologi Informasi.
- Satu kantor: Kantor PTPN I Regional 3 Semarang.
- Dua role: `USER` dan `ADMIN`.
- Satu jadwal kerja utama.
- Satu lokasi kerja utama.

## Alur Komunikasi

Mobile dan admin web akan berkomunikasi dengan backend melalui REST endpoint
di bawah `/api/v1`. Backend akan menjadi pemilik autentikasi, otorisasi,
aturan absensi, validasi geofence, laporan, dan audit log.

Untuk mobile pegawai:

```text
Flutter mobile
-> Golang REST API /api/v1/auth/*
-> PostgreSQL
```

Mobile menyimpan access token dan refresh token melalui secure storage.
Aplikasi mobile hanya mengizinkan profil backend dengan `role = USER` dan
`account_status = ACTIVE`. Akun admin tidak boleh dipakai untuk masuk aplikasi
pegawai.

Untuk basic attendance pegawai:

```text
Flutter mobile
-> Golang REST API /api/v1/attendance/*
-> PostgreSQL
```

Endpoint attendance dilindungi `Authenticate` dan `RequireRole(USER)`.
Backend membaca identitas pegawai dari access token, mengecek status akun di
database, menentukan tanggal kerja dengan timezone bisnis, lalu menyimpan
check-in/check-out sebagai `TIMESTAMPTZ`. Client tidak dipercaya untuk
mengirim user, role, tanggal, atau waktu.

Flutter mobile menampilkan status attendance dari endpoint tersebut melalui
repository dan `ChangeNotifier` berbasis Provider. Access token dan refresh
token tetap berasal dari sistem autentikasi mobile yang sama; tidak ada storage
token kedua. Jika endpoint attendance mengembalikan `401`, mobile melakukan
refresh token satu kali, menyimpan token rotasi, lalu mengulang request satu
kali.

Untuk admin web, browser tidak memanggil endpoint autentikasi Golang secara
langsung. Alur autentikasi admin adalah:

```text
Browser admin
-> Next.js Route Handler/BFF
-> Golang REST API
-> PostgreSQL
```

Next.js admin hanya menjadi BFF dan UI admin. Token dari Golang disimpan oleh
Next.js dalam HttpOnly cookie server-side. Token tidak boleh disimpan di
`localStorage` atau `sessionStorage`, dan Next.js tidak boleh mengakses
PostgreSQL secara langsung.

Untuk modul manajemen pegawai admin:

```text
Server Component /employees
-> server-only Golang API client
-> Golang REST API /api/v1/admin/employees
-> PostgreSQL
```

Create, update, dan perubahan status dari browser memakai BFF:

```text
Browser admin
-> Next.js Route Handler /api/admin/employees*
-> Golang REST API /api/v1/admin/employees*
-> PostgreSQL
```

Browser tidak menerima access token atau refresh token sebagai JSON. Token
tetap berada pada HttpOnly cookie dan hanya dibaca oleh server Next.js.

Untuk modul manajemen jadwal kerja dan assignment admin:

```text
Server Component /work-schedules dan /schedule-assignments
-> server-only Golang API client
-> Golang REST API /api/v1/admin/work-schedules
-> Golang REST API /api/v1/admin/schedule-assignments
-> PostgreSQL
```

Mutasi dari browser admin memakai BFF Next.js:

```text
Browser admin
-> Next.js Route Handler /api/admin/work-schedules*
-> Next.js Route Handler /api/admin/schedule-assignments*
-> Golang REST API /api/v1/admin/work-schedules*
-> Golang REST API /api/v1/admin/schedule-assignments*
-> PostgreSQL
```

Endpoint tersebut dilindungi `Authenticate` dan `RequireRole(ADMIN)`. Backend
memvalidasi jadwal dalam hari yang sama, status aktif schedule, role `USER`
untuk assignment, tanggal bisnis berbasis `BUSINESS_TIMEZONE`, serta konflik
periode assignment. PostgreSQL tetap menjadi pengaman terakhir untuk mencegah
assignment overlap saat ada request concurrent.

Browser admin tidak menerima token dari BFF. Access token dan refresh token
tetap berada dalam HttpOnly cookie yang hanya dibaca server Next.js. Assignment
yang dibuat admin menjadi sumber jadwal aktif bagi endpoint mobile attendance,
tetapi mobile tetap memakai Golang API langsung sebagai role `USER`.

## Prinsip Teknis

- Server time menjadi sumber waktu absensi yang otoritatif.
- Timezone bisnis default adalah `Asia/Jakarta` dan divalidasi saat aplikasi
  dimulai.
- Secret tidak boleh disimpan di repository.
- Perubahan schema database harus melalui migration.
- Error internal tidak boleh dikirim mentah ke client.
