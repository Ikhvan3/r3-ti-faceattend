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

## Prinsip Teknis

- Server time menjadi sumber waktu absensi yang otoritatif.
- Secret tidak boleh disimpan di repository.
- Perubahan schema database harus melalui migration.
- Error internal tidak boleh dikirim mentah ke client.
