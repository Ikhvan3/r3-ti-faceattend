# R3 TI FaceAttend Mobile

Aplikasi Flutter untuk pegawai Divisi Teknologi Informasi.

## Autentikasi Pegawai

Aplikasi hanya menerima akun backend dengan `role` bernilai `USER` dan
`account_status` bernilai `ACTIVE`. Akun `ADMIN`, `INACTIVE`, atau `SUSPENDED`
ditolak di sisi aplikasi mobile.

Endpoint backend yang dipakai:

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

Access token dan refresh token disimpan melalui `flutter_secure_storage`.
Aplikasi tidak mencetak token, password, atau header `Authorization`.

## Konfigurasi API

Default base URL:

```powershell
http://10.0.2.2:8080/api/v1
```

Jalankan di Android emulator:

```powershell
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

Jalankan di perangkat fisik melalui `adb reverse`:

```powershell
cd ..
.\scripts\adb-reverse.ps1
.\scripts\run-mobile.ps1
```

Jalankan di perangkat fisik melalui jaringan Wi-Fi:

```powershell
flutter run --dart-define=API_BASE_URL=http://<IP-LAPTOP>:8080/api/v1
```

Login mobile tidak membutuhkan token dari PowerShell. Aplikasi melakukan login,
memuat profil, menyimpan token, dan refresh session sendiri. Cleartext HTTP
hanya diizinkan untuk build debug Android melalui debug network security
config.

## Absensi Dasar

Setelah login sebagai pegawai `USER` aktif, beranda memanggil endpoint
attendance backend:

- `GET /api/v1/attendance/today`
- `POST /api/v1/attendance/check-in`
- `POST /api/v1/attendance/check-out`
- `GET /api/v1/attendance/history`

Check-in dan check-out memakai waktu server. Aplikasi tidak mengirim `user_id`,
tanggal absensi, waktu check-in, waktu check-out, timezone, atau role. Jika
access token kedaluwarsa, aplikasi melakukan refresh token dengan single-flight
agar beberapa request 401 paralel tidak memutar refresh token berkali-kali.

Check-in dan check-out memakai urutan GPS, liveness lokal, verifikasi wajah
backend, verification grant, lalu submit attendance. Koreksi absensi, cuti,
lembur, dan laporan admin belum tersedia pada tahap ini.

## Verifikasi

```powershell
dart format .
flutter analyze
flutter test
```
