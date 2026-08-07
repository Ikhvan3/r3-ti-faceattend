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

Untuk fondasi geofence lokasi kantor:

```text
Server Component /office-locations dan /location-assignments
-> server-only Golang API client
-> Golang REST API /api/v1/admin/office-locations
-> Golang REST API /api/v1/admin/location-assignments
-> PostgreSQL
```

Mutasi dari browser admin memakai BFF Next.js:

```text
Browser admin
-> Next.js Route Handler /api/admin/office-locations*
-> Next.js Route Handler /api/admin/location-assignments*
-> Golang REST API /api/v1/admin/office-locations*
-> Golang REST API /api/v1/admin/location-assignments*
-> PostgreSQL
```

Endpoint admin lokasi dilindungi `Authenticate` dan `RequireRole(ADMIN)`.
Backend memvalidasi koordinat, radius geofence, status aktif lokasi, role
`USER` untuk assignment, tanggal bisnis berbasis `BUSINESS_TIMEZONE`, dan
konflik periode assignment lokasi.

Pegawai mobile dapat membaca kebutuhan lokasi hari ini melalui:

```text
Flutter mobile
-> Golang REST API /api/v1/attendance/location-requirement
-> PostgreSQL
```

Endpoint tersebut dilindungi role `USER`, mengambil `user_id` dari token, dan
tidak menerima `user_id` dari client. Check-in/check-out mobile wajib mengirim
latitude, longitude, dan accuracy perangkat; backend memilih assignment lokasi
hari ini, menolak akurasi GPS yang melewati
`GEOFENCE_MAX_ACCURACY_METERS`, menghitung jarak Haversine server-side, dan
menyimpan evidence lokasi pada attendance record. Tidak ada API maps eksternal,
background location, atau tracking periodik. GPS spoofing tetap menjadi risiko
yang perlu mitigasi pada tahap berikutnya.

Untuk fondasi face enrollment:

```text
Flutter mobile atau client enrollment masa depan
-> Golang REST API /api/v1/face/*
-> PostgreSQL face_profiles
```

Backend mengambil identitas user dari access token role `USER`; request
enrollment tidak boleh membawa `user_id`. Golang hanya menerima dan menyimpan
embedding numerik yang sudah dibuat oleh komponen inferensi yang belum
diputuskan. Golang tidak menjalankan model ML dan tidak menambahkan dependensi
ML pada tahap ini.

Mobile memakai model TFLite `facenet.tflite` dari project open-source
`shubham0204/FaceRecognition_With_FaceNet_Android` berlisensi Apache-2.0.
Model identifier backend adalah `facenet`, version
`shubham0204-facenet-2020-fp32`, input tensor `[1,160,160,3]` `FLOAT32`, dan
output tensor `[1,128]` `FLOAT32`. Service face enrollment menolak
model/version yang tidak terdaftar, menolak dimensi selain `128`, dan menolak
nilai `NaN` atau `Inf`.

Pipeline Flutter:

```text
Camera preview
-> capture sample sementara
-> ML Kit face detection
-> quality checks
-> crop wajah + margin
-> resize 160x160
-> RGB float32 normalization (pixel - 127.5) / 127.5
-> TFLite FaceNet embedding
-> L2 normalize per sample
-> average 5 sample
-> L2 normalize hasil final
-> POST /api/v1/face/enroll
```

Response status/enrollment hanya menampilkan status, metadata model, dan waktu
enrollment. Embedding tidak dikembalikan ke client, tidak dicetak ke log, dan
tidak dipakai untuk attendance enforcement pada tahap ini.

Fondasi face verification standalone:

```text
Flutter mobile verification development flow
-> reuse ML Kit + TFLite embedding pipeline enrollment
-> POST /api/v1/face/verify
-> Golang mengambil face_profiles berdasarkan user JWT
-> cosine similarity
-> verified boolean
```

Candidate embedding menggunakan pipeline yang sama dengan enrollment:
orientation correction, detection, quality gate, crop margin, resize,
RGB normalization, inference, output validation, L2 normalization per sample,
average 5 sample, dan L2 normalization final. Backend menormalisasi ulang
candidate dan stored embedding sebelum cosine similarity agar kontrak normalized
embedding tetap aman.

Metric verifikasi adalah cosine similarity dengan range valid `[-1,1]`.
Threshold berada di backend melalui `FACE_VERIFICATION_THRESHOLD` dan divalidasi
saat startup. Repository ini belum memiliki bukti kalibrasi threshold yang cukup,
sehingga tidak ada default produksi; nilai development wajib diisi dan harus
dikalibrasi dari pengujian enrollment/verifikasi user lokal. Verification masih
standalone, belum menjadi syarat check-in/check-out, dan belum memiliki liveness
atau anti-spoofing.

## Prinsip Teknis

- Server time menjadi sumber waktu absensi yang otoritatif.
- Timezone bisnis default adalah `Asia/Jakarta` dan divalidasi saat aplikasi
  dimulai.
- Secret tidak boleh disimpan di repository.
- Perubahan schema database harus melalui migration.
- Error internal tidak boleh dikirim mentah ke client.
