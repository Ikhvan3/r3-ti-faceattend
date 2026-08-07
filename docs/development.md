# Development

Dokumen ini mencatat cara kerja lokal untuk repository R3 TI FaceAttend.

## Prasyarat

- Git.
- Flutter dan Android toolchain.
- Node.js dan npm.
- Golang.
- PostgreSQL lokal.

## Mobile

```powershell
cd mobile
flutter pub get
dart format .
flutter analyze
flutter test
```

Konfigurasi awal dicatat di `mobile/.env.example`. Pada tahap berikutnya nilai
runtime dapat diberikan melalui `--dart-define`.

Autentikasi mobile menggunakan Golang API langsung. Default lokal untuk Android
emulator:

```powershell
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

Untuk perangkat fisik melalui `adb reverse`:

```powershell
adb reverse tcp:8080 tcp:8080
flutter run --dart-define=API_BASE_URL=http://127.0.0.1:8080/api/v1
```

Untuk perangkat fisik melalui Wi-Fi, ganti host dengan IP laptop lokal.
Gunakan akun dummy `USER` aktif dari endpoint admin employee. Jangan menyimpan
credential, token, atau header `Authorization` pada log atau dokumentasi.

Alur uji manual mobile:

1. Jalankan Golang API dan pastikan `/api/v1/health` `ok`.
2. Pastikan ada pegawai dummy dengan `role = USER` dan `account_status = ACTIVE`.
3. Jalankan aplikasi dengan `API_BASE_URL` yang sesuai device.
4. Login sebagai pegawai dummy aktif.
5. Tutup dan buka ulang aplikasi untuk memastikan restore session.
6. Paksa access token expired lalu pastikan refresh berjalan satu kali.
7. Logout dan pastikan kembali ke halaman login.
8. Coba akun admin dan akun nonaktif untuk memastikan akses ditolak.

Alur uji manual mobile attendance pada HP Android melalui USB:

```powershell
$adb = "$env:LOCALAPPDATA\Android\Sdk\platform-tools\adb.exe"
& $adb reverse tcp:8080 tcp:8080

cd mobile
flutter run -d DEVICE_ID `
  --dart-define="API_BASE_URL=http://127.0.0.1:8080/api/v1"
```

Pastikan backend aktif, migration attendance dan lokasi sudah dijalankan, dan
pegawai dummy `USER` aktif sudah memiliki assignment jadwal serta assignment
lokasi untuk tanggal bisnis hari ini. Aktifkan GPS perangkat. Setelah login,
beranda menampilkan status attendance hari ini, tombol check-in/check-out
sesuai state, dan halaman riwayat memakai pagination backend. Check-in dan
check-out mengambil satu lokasi terbaru dari perangkat, mengirim latitude,
longitude, dan accuracy ke Golang API, lalu backend melakukan enforcement
geofence server-side. Kamera, verifikasi wajah, liveness, koreksi absensi,
cuti, lembur, dan laporan admin belum tersedia pada tahap ini.

Verifikasi otomatis mobile:

```powershell
cd mobile
dart format .
flutter analyze
flutter test
```

## Admin Web

```powershell
cd admin-web
npm install
npm run lint
npx tsc --noEmit
npm run dev
```

Konfigurasi awal dicatat di `admin-web/.env.example`.

Environment lokal admin web:

```powershell
cd admin-web
$env:APP_ENV="local"
$env:GO_API_BASE_URL="http://127.0.0.1:8080/api/v1"
npm run dev
```

Gunakan `http://localhost:3000/login` untuk login admin. Browser admin
memanggil Next.js Route Handler di `/api/auth/*`; browser tidak memanggil
endpoint autentikasi Golang secara langsung.

Next.js menyimpan access token dan refresh token dalam HttpOnly cookie:

- `r3_access_token`
- `r3_refresh_token`

Pada `APP_ENV=local`, cookie memakai `secure=false` agar bisa berjalan melalui
HTTP lokal. Pada `APP_ENV=production`, cookie memakai `secure=true`.

Token tidak boleh disimpan di `localStorage` atau `sessionStorage`.

Alur uji login/logout manual:

1. Jalankan Golang API dengan konfigurasi auth.
2. Pastikan migration dan admin seed sudah tersedia.
3. Jalankan admin web dengan `GO_API_BASE_URL`.
4. Buka `http://localhost:3000/login`.
5. Login dengan akun admin seed.
6. Pastikan diarahkan ke `/dashboard`.
7. Pastikan profil admin tampil.
8. Pastikan `localStorage` dan `sessionStorage` tidak berisi token.
9. Pastikan cookie `r3_access_token` dan `r3_refresh_token` berstatus HttpOnly.
10. Klik logout dan pastikan kembali ke `/login`.
11. Coba password salah dan pastikan pesan error tetap generik.
12. Matikan Golang API dan pastikan UI menampilkan error aman.

### Admin Web Employee Management

Route halaman admin employee:

- `GET /employees`
- `GET /employees/new`
- `GET /employees/{id}`
- `GET /employees/{id}/edit`

BFF endpoint yang dipanggil browser admin:

- `POST /api/admin/employees`
- `PUT /api/admin/employees/{id}`
- `PATCH /api/admin/employees/{id}/status`

Server Components membaca list dan detail pegawai langsung dari Golang melalui
server-only API client dengan `GO_API_BASE_URL`. Browser tidak memanggil Golang
secara langsung dan tidak menerima token sebagai JSON.

Jalankan backend dan admin web:

```powershell
# Terminal 1
cd backend
go run ./cmd/api

# Terminal 2
cd admin-web
$env:APP_ENV="local"
$env:GO_API_BASE_URL="http://127.0.0.1:8080/api/v1"
npm run dev
```

Alur uji CRUD pegawai admin:

1. Login admin di `http://localhost:3000/login`.
2. Buka `http://localhost:3000/employees`.
3. Pastikan daftar pegawai tampil atau empty state aman.
4. Gunakan search dan filter status; pastikan query muncul di URL.
5. Tambahkan pegawai dummy dari `/employees/new`.
6. Coba email duplikat dan pastikan pesan conflict tampil.
7. Coba nomor pegawai duplikat dan pastikan pesan conflict tampil.
8. Buka detail pegawai.
9. Edit data pegawai.
10. Ubah status ke `INACTIVE`, `SUSPENDED`, lalu `ACTIVE` sesuai kebutuhan uji.
11. Pastikan `password_hash` tidak terlihat di UI atau response BFF.
12. Pastikan `localStorage` dan `sessionStorage` tidak menyimpan token.
13. Hapus/ubah cookie session untuk memastikan session invalid diarahkan ke
    login atau menampilkan pesan aman.

Gunakan hanya data dummy development. Jangan memakai nama, email, nomor
pegawai, nomor telepon, atau data personal PTPN yang sebenarnya pada database
lokal, dokumentasi, commit, issue, screenshot, atau log.

### Admin Web Jadwal Kerja dan Penugasan

Route halaman admin schedule:

- `GET /work-schedules`
- `GET /work-schedules/new`
- `GET /work-schedules/{id}`
- `GET /work-schedules/{id}/edit`
- `GET /schedule-assignments`
- `GET /schedule-assignments/new`
- `GET /schedule-assignments/{id}`

BFF endpoint yang dipanggil browser admin:

- `POST /api/admin/work-schedules`
- `PUT /api/admin/work-schedules/{id}`
- `PATCH /api/admin/work-schedules/{id}/status`
- `POST /api/admin/schedule-assignments`
- `PATCH /api/admin/schedule-assignments/{id}/end`

Server Components membaca list dan detail langsung dari Golang REST API dengan
token HttpOnly cookie server-side. Browser hanya mengirim mutasi ke BFF Next.js
dan tidak menerima access token atau refresh token sebagai JSON.

Alur create schedule:

1. Admin membuka `/work-schedules/new`.
2. Form mengirim `name`, `start_time`, `end_time`, dan `grace_minutes` ke BFF.
3. BFF menolak field tambahan seperti `is_active`, `created_at`, dan
   `updated_at`, lalu meneruskan request ke Golang API.
4. Jadwal baru dibuat aktif oleh backend.

Alur create dan end assignment:

1. Admin membuka `/schedule-assignments/new`.
2. Pilihan pegawai memakai `user_id` dari pegawai `USER`; pilihan jadwal hanya
   memakai jadwal aktif.
3. Form mengirim `user_id`, `schedule_id`, `effective_from`, dan
   `effective_to` opsional dalam format `YYYY-MM-DD`.
4. Jika periode overlap, backend mengembalikan `409` dan UI menampilkan pesan
   aman.
5. Detail assignment menyediakan tombol `Akhiri Penugasan` untuk assignment
   yang belum berakhir. Tanggal akhir bersifat inklusif.

Assignment jadwal menjadi sumber jadwal untuk endpoint mobile attendance.
Mobile tetap membaca jadwal melalui Golang API attendance sebagai role `USER`;
admin web tidak mengubah alur token mobile, attendance record lama, GPS,
geofence, kamera, face recognition, liveness, atau laporan.

Gunakan hanya data dummy development. Jangan memakai data pegawai PTPN yang
sebenarnya pada database lokal, dokumentasi, commit, issue, screenshot, atau
log.

### Admin Web Lokasi Kantor dan Penugasan Lokasi

Route halaman admin lokasi:

- `GET /office-locations`
- `GET /office-locations/new`
- `GET /office-locations/{id}`
- `GET /office-locations/{id}/edit`
- `GET /location-assignments`
- `GET /location-assignments/new`
- `GET /location-assignments/{id}`

BFF endpoint yang dipanggil browser admin:

- `POST /api/admin/office-locations`
- `PUT /api/admin/office-locations/{id}`
- `PATCH /api/admin/office-locations/{id}/status`
- `POST /api/admin/location-assignments`
- `PATCH /api/admin/location-assignments/{id}/end`

Server Components membaca list dan detail langsung dari Golang REST API dengan
token HttpOnly cookie server-side. Browser hanya mengirim mutasi ke BFF Next.js
dan tidak menerima access token atau refresh token sebagai JSON.

Alur create lokasi:

1. Admin membuka `/office-locations/new`.
2. Form mengirim `name`, `address`, `latitude`, `longitude`, dan
   `radius_meters` ke BFF.
3. BFF menolak field tambahan seperti `is_active`, `created_at`, dan
   `updated_at`, lalu meneruskan request ke Golang API.
4. Backend membuat lokasi baru sebagai lokasi aktif.

Alur create dan end penugasan lokasi:

1. Admin membuka `/location-assignments/new`.
2. Pilihan pegawai memakai pegawai `USER`; pilihan lokasi hanya memakai lokasi
   aktif.
3. Form mengirim `user_id`, `office_location_id`, `effective_from`, dan
   `effective_to` opsional dalam format `YYYY-MM-DD`.
4. Jika periode overlap, backend mengembalikan `409` dan UI menampilkan pesan
   aman.
5. Detail assignment menyediakan tombol `Akhiri Penugasan` untuk assignment
   yang belum berakhir. Tanggal akhir bersifat inklusif.

Data lokasi dan penugasan lokasi menjadi sumber enforcement geofence untuk
check-in/check-out mobile. Mobile tidak melakukan polling GPS, tidak menyimpan
riwayat koordinat lokal, dan tidak mengirim `user_id`, `location_id`, jarak,
atau timestamp perangkat.

## Backend

```powershell
cd backend
go test ./...
go vet ./...
go run ./cmd/api
```

Konfigurasi awal dicatat di `backend/.env.example`.

### Menyiapkan Database Lokal

Buat database PostgreSQL lokal sebelum menjalankan API:

```powershell
createdb -U postgres r3_ti_faceattend
```

Jika `createdb` tidak tersedia, gunakan `psql`:

```powershell
psql -U postgres -c "CREATE DATABASE r3_ti_faceattend;"
```

### Konfigurasi Environment Backend

Backend membaca konfigurasi dari environment variable proses. File
`backend/.env.example` hanya menjadi contoh nilai lokal, bukan file yang
dibaca otomatis oleh aplikasi.

Variabel yang digunakan backend:

- `APP_ENV`
- `APP_PORT`
- `BUSINESS_TIMEZONE`
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_SSLMODE`
- `AUTH_ACCESS_TOKEN_SECRET`
- `AUTH_ACCESS_TOKEN_TTL_MINUTES`
- `AUTH_REFRESH_TOKEN_TTL_HOURS`
- `AUTH_TOKEN_ISSUER`
- `AUTH_TOKEN_AUDIENCE`
- `GEOFENCE_MAX_ACCURACY_METERS`

PowerShell dapat memuat nilai untuk sesi terminal seperti ini:

```powershell
$env:APP_ENV="local"
$env:APP_PORT="8080"
$env:BUSINESS_TIMEZONE="Asia/Jakarta"
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_NAME="r3_ti_faceattend"
$env:DB_USER="postgres"
$env:DB_PASSWORD=
$env:DB_SSLMODE="disable"
$env:AUTH_ACCESS_TOKEN_SECRET="dKIkuNUj695Bwr04mMc8P7QTSvCpeaFxAHhXWEyqsJzRVgD1"
$env:AUTH_ACCESS_TOKEN_TTL_MINUTES="15"
$env:AUTH_REFRESH_TOKEN_TTL_HOURS="168"
$env:AUTH_TOKEN_ISSUER="r3-ti-faceattend-api"
$env:AUTH_TOKEN_AUDIENCE="r3-ti-faceattend-client"
$env:GEOFENCE_MAX_ACCURACY_METERS="50"
go run ./cmd/api
```

`AUTH_ACCESS_TOKEN_SECRET` wajib diisi dan tidak boleh dicetak ke log. TTL
access token dan refresh token harus bernilai positif.
`BUSINESS_TIMEZONE` default `Asia/Jakarta` dan harus bernilai timezone IANA
yang valid. `GEOFENCE_MAX_ACCURACY_METERS` default `50` dan harus berupa angka
finite lebih besar dari `0`.

### Seed Admin Lokal

Seeder admin hanya digunakan untuk local development setelah migration database
sudah dijalankan. Command ini tidak menjalankan migration otomatis, tidak
membuat tabel, dan tidak boleh dipakai untuk menyimpan credential di repository.

Password admin development minimal 8 karakter. Password plaintext hanya dibaca
dari environment, di-hash dengan bcrypt, dan tidak dicetak ke output.

PowerShell:

```powershell
cd backend
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_NAME="r3_ti_faceattend"
$env:DB_USER="postgres"
$env:DB_PASSWORD="isi-di-sesi-lokal"
$env:DB_SSLMODE="disable"

$env:ADMIN_EMPLOYEE_NUMBER="ADMIN-LOCAL"
$env:ADMIN_NAME="Admin Lokal"
$env:ADMIN_EMAIL="admin.local@example.test"
$env:ADMIN_PASSWORD="minimal-8-karakter"
$env:ADMIN_POSITION="Administrator TI"

go run ./cmd/seed-admin
```

Jalankan command yang sama untuk kedua kali untuk memastikan seed bersifat
idempotent. Jika admin yang sama sudah ada, command selesai tanpa membuat
duplikasi dan tanpa mengubah password.

Verifikasi aman melalui `psql` tanpa menampilkan hash penuh:

```powershell
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "SELECT employee_number, name, email, role, account_status, created_at, password_hash <> '' AS password_hash_filled FROM users WHERE role = 'ADMIN';"
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "SELECT COUNT(*) AS admin_count FROM users WHERE role = 'ADMIN' AND lower(email) = lower('admin.local@example.test');"
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "SELECT password_hash = 'minimal-8-karakter' AS password_is_plaintext FROM users WHERE lower(email) = lower('admin.local@example.test');"
```

Bersihkan password dari environment setelah selesai:

```powershell
Remove-Item Env:ADMIN_PASSWORD
```

Jangan commit credential, password contoh nyata, atau file environment lokal
yang berisi secret.

### Autentikasi Backend

Pastikan migration `000001` dan `000002` sudah dijalankan sebelum mencoba
login. Admin development juga perlu dibuat dengan `go run ./cmd/seed-admin`.

Jalankan API:

```powershell
cd backend
go run ./cmd/api
```

Login:

```powershell
$LoginBody = @{
  email = $env:ADMIN_EMAIL
  password = $env:ADMIN_PASSWORD
} | ConvertTo-Json

$Login = Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/auth/login" `
  -ContentType "application/json" `
  -Body $LoginBody
```

Gunakan access token sebagai Bearer token:

```powershell
$Headers = @{
  Authorization = "Bearer $($Login.data.access_token)"
}

Invoke-RestMethod -Method Get -Uri "http://localhost:8080/api/v1/auth/me" -Headers $Headers
Invoke-RestMethod -Method Get -Uri "http://localhost:8080/api/v1/admin/ping" -Headers $Headers
```

Refresh:

```powershell
$RefreshBody = @{
  refresh_token = $Login.data.refresh_token
} | ConvertTo-Json

$Refresh = Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/auth/refresh" `
  -ContentType "application/json" `
  -Body $RefreshBody
```

Logout:

```powershell
$LogoutBody = @{
  refresh_token = $Refresh.data.refresh_token
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/auth/logout" `
  -ContentType "application/json" `
  -Body $LogoutBody
```

Jangan menulis token atau secret nyata ke dokumentasi, commit, atau log.

### Employee Management Backend

Employee management backend berjalan di bawah route admin Golang:

- `GET /api/v1/admin/employees`
- `POST /api/v1/admin/employees`
- `GET /api/v1/admin/employees/{id}`
- `PUT /api/v1/admin/employees/{id}`
- `PATCH /api/v1/admin/employees/{id}/status`

Semua route membutuhkan Bearer access token role `ADMIN`. Gunakan hanya data
dummy development. Jangan memakai nama, email, nomor pegawai, nomor telepon,
atau data personal PTPN yang sebenarnya pada database lokal, dokumentasi,
commit, issue, screenshot, atau log.

Alur uji manual dengan PostgreSQL lokal:

1. Login sebagai admin development melalui `/api/v1/auth/login`.
2. Simpan access token ke `$Headers`.
3. Jalankan list employee dan pastikan response aman.
4. Tambahkan pegawai dummy dengan email domain contoh seperti
   `example.test`.
5. Baca detail pegawai dari `id` response create.
6. Update profil pegawai dummy.
7. Ubah status ke `INACTIVE`.
8. Coba create dengan email yang sama dan pastikan HTTP `409`.
9. Coba create dengan nomor pegawai yang sama dan pastikan HTTP `409`.
10. Pastikan response tidak mengandung `password_hash`.
11. Login sebagai role `USER` dummy dan pastikan route admin mengembalikan
    HTTP `403`.

Contoh PowerShell lengkap ada di `docs/api.md`.

### Basic Attendance Backend

Jalankan migration sampai `000003`, lalu siapkan pegawai dummy `USER` aktif.
Seeder attendance development hanya boleh dijalankan dengan
`APP_ENV=development`; command tidak menyimpan credential dan tidak berjalan
otomatis di production.

```powershell
cd backend
$env:APP_ENV="development"
$env:BUSINESS_TIMEZONE="Asia/Jakarta"
$env:DEV_ATTENDANCE_USER_EMAIL="pegawai.dummy.ti@example.test"
$env:DEV_ATTENDANCE_SCHEDULE_NAME="Jadwal Kerja Dummy TI"
$env:DEV_ATTENDANCE_START_TIME="08:00"
$env:DEV_ATTENDANCE_END_TIME="17:00"
$env:DEV_ATTENDANCE_GRACE_MINUTES="15"
go run ./cmd/seed-attendance-dev
```

Setelah backend admin schedule tersedia, assignment development lebih baik
dibuat melalui endpoint admin:

1. Login sebagai admin dummy.
2. Buat jadwal melalui `POST /api/v1/admin/work-schedules`.
3. Buat assignment pegawai dummy melalui
   `POST /api/v1/admin/schedule-assignments`.
4. Jalankan login mobile sebagai pegawai dummy dan baca
   `GET /api/v1/attendance/today`.

Seeder `seed-attendance-dev` tetap hanya alat development lokal untuk
menyiapkan data dummy cepat. Jangan gunakan nama, email, nomor pegawai, nomor
telepon, atau data personal PTPN yang sebenarnya.

### Admin Work Schedule Backend

Endpoint backend:

- `GET /api/v1/admin/work-schedules`
- `POST /api/v1/admin/work-schedules`
- `GET /api/v1/admin/work-schedules/{id}`
- `PUT /api/v1/admin/work-schedules/{id}`
- `PATCH /api/v1/admin/work-schedules/{id}/status`

Endpoint assignment:

- `GET /api/v1/admin/schedule-assignments`
- `POST /api/v1/admin/schedule-assignments`
- `GET /api/v1/admin/schedule-assignments/{id}`
- `PATCH /api/v1/admin/schedule-assignments/{id}/end`

Semua endpoint membutuhkan token role `ADMIN`. Token role `USER` harus
mengembalikan HTTP `403`. Query pagination dijalankan di PostgreSQL dengan
`page_size` maksimum `100`.

Alur uji manual dengan PostgreSQL lokal:

1. Jalankan migration sampai `000004`.
2. Login sebagai admin dummy dan simpan Bearer token ke `$Headers`.
3. Buat satu jadwal aktif.
4. List jadwal dan pastikan data muncul.
5. Update jadwal.
6. Buat pegawai dummy `USER` aktif melalui endpoint employee bila belum ada.
7. Buat assignment ke pegawai dummy.
8. Pastikan assignment muncul pada list.
9. Coba assignment overlap dan pastikan HTTP `409`.
10. Buat assignment berurutan mulai tanggal setelah `effective_to` sebelumnya.
11. Pastikan jadwal dengan assignment aktif atau masa depan tidak dapat
    dinonaktifkan.
12. Akhiri assignment dengan tanggal hari ini atau tanggal setelah
    `effective_from`; tanggal akhir bersifat inklusif.
13. Pastikan endpoint attendance `USER` tetap dapat membaca jadwal aktifnya.
14. Login sebagai token `USER` dan pastikan seluruh route admin HTTP `403`.
15. Jalankan list pada filter yang kosong dan pastikan response `items: []`.
16. Uji migration down/up pada database development yang aman bila memungkinkan.

Endpoint USER:

- `GET /api/v1/attendance/today`
- `POST /api/v1/attendance/check-in`
- `POST /api/v1/attendance/check-out`
- `GET /api/v1/attendance/history?page=1&page_size=10`

Alur uji manual dengan PostgreSQL lokal:

1. Login sebagai pegawai dummy `USER` aktif melalui `/api/v1/auth/login`.
2. Simpan access token ke header Bearer.
3. `GET /api/v1/attendance/today` sebelum check-in harus menampilkan
   `NOT_CHECKED_IN`.
4. `POST /api/v1/attendance/check-in` dengan body latitude, longitude, dan
   `accuracy_meters` harus HTTP `201`.
5. Check-in kedua harus HTTP `409`.
6. `GET /api/v1/attendance/today` harus menampilkan `CHECKED_IN`.
7. `POST /api/v1/attendance/check-out` dengan body latitude, longitude, dan
   `accuracy_meters` harus HTTP `200`.
8. Check-out kedua harus HTTP `409`.
9. `GET /api/v1/attendance/today` harus menampilkan `COMPLETED`.
10. `GET /api/v1/attendance/history` harus menampilkan record user tersebut.
11. Kolom evidence `check_in_location_id`, `check_in_latitude`,
    `check_in_longitude`, `check_in_accuracy_meters`,
    `check_in_distance_meters`, dan pasangan check-out harus terisi.
12. Token `ADMIN` harus ditolak HTTP `403`.
13. Token `USER` lain tidak dapat melihat record user tersebut.

### Office Location Backend

Endpoint admin lokasi:

- `GET /api/v1/admin/office-locations`
- `POST /api/v1/admin/office-locations`
- `GET /api/v1/admin/office-locations/{id}`
- `PUT /api/v1/admin/office-locations/{id}`
- `PATCH /api/v1/admin/office-locations/{id}/status`

Endpoint assignment lokasi:

- `GET /api/v1/admin/location-assignments`
- `POST /api/v1/admin/location-assignments`
- `GET /api/v1/admin/location-assignments/{id}`
- `PATCH /api/v1/admin/location-assignments/{id}/end`

Endpoint user:

- `GET /api/v1/attendance/location-requirement`

Semua route admin membutuhkan token role `ADMIN`. Endpoint
`location-requirement` hanya menerima token role `USER` dan mengambil identitas
pegawai dari token. Check-in/check-out kini wajib menerima koordinat perangkat
dan melakukan enforcement geofence di backend.

Alur uji manual dengan PostgreSQL lokal:

1. Jalankan migration sampai `000005`.
2. Login sebagai admin dummy dan simpan Bearer token ke `$Headers`.
3. Buat lokasi aktif melalui `POST /api/v1/admin/office-locations`.
4. List lokasi dan pastikan data muncul.
5. Update `radius_meters`.
6. Buat pegawai dummy `USER` aktif bila belum ada.
7. Buat assignment lokasi kepada USER dummy.
8. Login sebagai USER dummy dan panggil
   `GET /api/v1/attendance/location-requirement`.
9. Coba assignment overlap dan pastikan HTTP `409`.
10. Buat assignment berurutan mulai tanggal setelah `effective_to` sebelumnya.
11. Pastikan lokasi dengan assignment aktif atau masa depan tidak dapat
    dinonaktifkan.
12. Akhiri assignment dengan tanggal hari ini atau setelah `effective_from`.
13. Pastikan lokasi tanpa assignment aktif dapat dinonaktifkan.
14. Token `USER` harus ditolak dari route admin lokasi.
15. Token `ADMIN` harus ditolak dari endpoint `location-requirement`.
16. Pastikan `POST /api/v1/attendance/check-in` dan
    `POST /api/v1/attendance/check-out` dengan body koordinat di dalam radius
    tetap bekerja.
17. Pastikan record attendance lama tetap terbaca dari `today` dan `history`.

Keterbatasan: fondasi ini belum mendeteksi GPS spoofing, belum memakai kamera,
belum melakukan face recognition/liveness, dan belum melakukan background
location atau tracking periodik.

### Face Enrollment Foundation Backend

Endpoint user:

- `GET /api/v1/face/status`
- `POST /api/v1/face/enroll`
- `DELETE /api/v1/face/enrollment`

Alur uji manual dengan PostgreSQL lokal:

1. Jalankan migration sampai `000006`.
2. Login sebagai pegawai dummy `USER` dengan `account_status` `ACTIVE`.
3. Panggil `GET /api/v1/face/status`; user tanpa profile harus mendapat
   `NOT_ENROLLED`.
4. Panggil `POST /api/v1/face/enroll` dengan `embedding`,
   `embedding_model`, dan `embedding_version` sesuai registry model backend.
5. Pastikan response enrollment tidak memuat field `embedding` atau nilai
   embedding.
6. Panggil `GET /api/v1/face/status`; user enrolled harus mendapat
   `ENROLLED`, `embedding_model`, `embedding_version`, dan `enrolled_at`.
7. Enrollment kedua tanpa reset harus HTTP `409`.
8. Panggil `DELETE /api/v1/face/enrollment`.
9. Panggil `GET /api/v1/face/status`; status harus kembali `NOT_ENROLLED`.
10. Enrollment ulang setelah reset harus diizinkan.
11. Token `ADMIN` harus ditolak HTTP `403`.
12. Request dengan `user_id` di body harus ditolak sebagai malformed request.
13. Embedding kosong, nilai `NaN`/`Inf`, model tidak didukung, atau dimensi
    tidak sesuai harus ditolak.

Verifikasi pgvector bila kredensial database lokal tersedia:

```powershell
psql -h localhost -p 5432 -U postgres -d r3_ti_faceattend -c "SELECT extname FROM pg_extension WHERE extname = 'vector';"
```

Model enrollment mobile:

- Model identifier: `facenet`
- Version: `shubham0204-facenet-2020-fp32`
- File: `mobile/assets/models/facenet.tflite`
- SHA-256:
  `D7C1F7F130376982C7004920DDC41925AC2E5AECF6522F476C8BBB3669DB7013`
- Source: `shubham0204/FaceRecognition_With_FaceNet_Android`
- License: Apache-2.0, disimpan di
  `mobile/assets/models/facenet.APACHE-2.0.txt`
- Input tensor nyata: `[1,160,160,3]` `FLOAT32`
- Output tensor nyata: `[1,128]` `FLOAT32`

Preprocessing Flutter:

1. Ambil capture sementara dari kamera depan.
2. Deteksi wajah dengan ML Kit.
3. Terima hanya tepat satu wajah.
4. Tolak wajah terlalu kecil, terlalu dekat tepi frame, atau pose terlalu
   menyamping/miring.
5. Bake orientation gambar.
6. Crop bounding box wajah dengan margin.
7. Resize ke `160x160`.
8. Gunakan channel RGB.
9. Normalisasi pixel ke `(pixel - 127.5) / 127.5`.
10. Jalankan TFLite.
11. Validasi output finite dan berdimensi `128`.
12. L2 normalize setiap sample.
13. Ambil target awal `5` sample dengan interval.
14. Average element-wise dan L2 normalize hasil final.

Pada tahap ini migration tetap memakai `DOUBLE PRECISION[]`, bukan `vector(n)`.
Setelah kebutuhan similarity search backend jelas, pertimbangkan migration
lanjutan ke `pgvector`.

Keterbatasan: endpoint dan mobile enrollment ini belum melakukan face
verification pada attendance, belum liveness, belum anti-spoofing, dan
enrollment belum diwajibkan saat check-in/check-out.

### Face Verification Foundation

Backend:

1. Isi `FACE_VERIFICATION_THRESHOLD` di environment backend. Nilai ini wajib,
   harus finite, dan berada dalam range cosine similarity `[-1,1]`.
2. Jalankan API dengan user development yang sudah memiliki row `face_profiles`
   berstatus `ENROLLED`.
3. Panggil `GET /api/v1/face/status` memakai JWT USER dan pastikan response
   `ENROLLED`.
4. Panggil `POST /api/v1/face/verify` dengan `embedding`,
   `embedding_model`, dan `embedding_version` dari pipeline Flutter yang sama
   dengan enrollment.
5. Pastikan response hanya memuat `verified`, bukan embedding, threshold,
   similarity score, SQL detail, atau data internal.
6. Candidate mismatch tetap harus menghasilkan HTTP `200` dengan
   `verified: false`.
7. Wrong dimension, wrong model/version, `NaN`, `Inf`, zero vector, unknown
   field, atau malformed JSON harus ditolak.

Kalibrasi threshold:

- Metric saat ini adalah cosine similarity.
- Source model FaceNet Android menggunakan embedding 128 dimensi dan membahas
  cosine similarity/L2 distance untuk membandingkan embedding.
- Repository ini belum menyimpan hasil pengujian threshold yang cukup untuk
  default yang dapat dipertanggungjawabkan.
- Gunakan data development dari user yang sama yang sudah enrollment, lalu uji
  beberapa capture valid orang yang sama dan beberapa dummy/candidate berbeda.
  Pilih threshold setelah mengamati false rejection dan false acceptance.
- Flutter tidak boleh mengatur, mengetahui, atau mengirim threshold.

Flutter:

1. Login sebagai USER.
2. Pastikan Home face card menampilkan `Terdaftar`.
3. Tekan `Uji Verifikasi Wajah`.
4. Kamera depan harus tampil.
5. Flow verification memakai quality gate enrollment: exactly one face, ukuran
   cukup, posisi dalam area, pose diterima, dan frame valid.
6. Candidate embedding hanya berada di memory selama proses verifikasi dan
   dikirim ke backend.
7. Hasil `verified: true` tampil sebagai `Wajah berhasil diverifikasi.`
8. Hasil `verified: false` tampil sebagai
   `Wajah tidak cocok dengan data yang terdaftar.`
9. Kegagalan verification tidak melakukan logout.
10. Attendance/geofence lama tetap diuji terpisah karena verification belum
    menjadi syarat attendance.

### Active Face Liveness Prototype

Flutter menyediakan halaman standalone `Uji Keaktifan Wajah` dari Home face
card ketika status wajah sudah `ENROLLED`. Halaman ini belum terhubung ke
attendance dan tidak mengubah geofence.

Flow:

1. Kamera depan menampilkan preview dengan aspect ratio native.
2. ML Kit memproses `CameraImage` stream dengan classification dan tracking.
3. Challenge acak 2-3 langkah dimulai saat tepat satu wajah valid berada di
   tengah frame.
4. Aksi minimum: blink, hadap sedikit ke kiri, hadap sedikit ke kanan, dan
   kembali menghadap depan.
5. Blink harus `OPEN -> CLOSED -> OPEN`; satu frame mata tertutup tidak cukup.
6. Head turn memakai `headEulerAngleY`; mapping kiri/kanan melewati helper
   orientation/mirroring dan wajib dikalibrasi pada HP fisik.
7. Setelah liveness pass, stream dihentikan, aplikasi mengambil sample
   sementara, memakai ulang pipeline face verification, lalu memanggil
   `POST /api/v1/face/verify`.
8. Raw image, crop wajah, screenshot, base64 image, dan candidate embedding
   tidak disimpan sebagai data aplikasi dan tidak boleh dicetak ke log.

Threshold development berada di `LivenessConfig`: eye open/closed, yaw turn,
yaw center, roll, ukuran wajah, margin tepi, timeout per action, timeout total,
dan timeout face lost. Nilai tersebut adalah titik awal development dan harus
dikalibrasi dari HP fisik yang dipakai.

Alur uji manual HP Android:

1. Login sebagai `USER`.
2. Pastikan status wajah `ENROLLED`.
3. Buka `Uji Keaktifan Wajah`.
4. Pastikan preview tidak gepeng.
5. Tepat satu wajah dapat memulai challenge.
6. Tanpa wajah ditolak.
7. Dua wajah ditolak.
8. Blink terdeteksi hanya setelah open, closed, lalu open.
9. Turn left terdeteksi.
10. Turn right terdeteksi.
11. Challenge berubah antar sesi.
12. Wajah statis tanpa aksi tidak pass.
13. Timeout bekerja.
14. Liveness pass.
15. Face verification setelah liveness menghasilkan `verified=true` untuk user
    yang sama.
16. Retry berhasil setelah gagal.
17. Home tetap berfungsi.
18. Geofence attendance lama tetap bekerja.
19. Tidak ada token, image, atau embedding di log.

Verifikasi otomatis mobile:

```powershell
cd mobile
dart format .
flutter analyze
flutter test
```

Unit test liveness memakai fake frame/detector dan tidak membutuhkan kamera.
Coverage mencakup zero face, multiple face, invalid quality, tracking berubah,
blink valid, mata tertutup tanpa reopen, head turn, return center, random
challenge, timeout, face lost, liveness completed, hasil verify true/false, API
error, retry, dan pencegahan multiple submit.

### Health Check

Endpoint health tersedia di:

- `GET /health`
- `GET /api/v1/health`

Contoh verifikasi:

```powershell
curl http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/api/v1/health
```

Jika API hidup tetapi database belum tersambung, response tetap JSON konsisten
dengan status `degraded` dan HTTP status `503`.

## Menjalankan Flutter pada HP Android melalui USB

Pastikan backend berjalan pada port 8080.

```powershell
$adb = "$env:LOCALAPPDATA\Android\Sdk\platform-tools\adb.exe"

& $adb devices
& $adb reverse tcp:8080 tcp:8080
& $adb reverse --list

flutter run -d DEVICE_ID `
  --dart-define="API_BASE_URL=http://127.0.0.1:8080/api/v1"


