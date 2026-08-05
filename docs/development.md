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
go run ./cmd/api
```

`AUTH_ACCESS_TOKEN_SECRET` wajib diisi dan tidak boleh dicetak ke log. TTL
access token dan refresh token harus bernilai positif.
`BUSINESS_TIMEZONE` default `Asia/Jakarta` dan harus bernilai timezone IANA
yang valid.

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
4. `POST /api/v1/attendance/check-in` dengan body kosong harus HTTP `201`.
5. Check-in kedua harus HTTP `409`.
6. `GET /api/v1/attendance/today` harus menampilkan `CHECKED_IN`.
7. `POST /api/v1/attendance/check-out` dengan body kosong harus HTTP `200`.
8. Check-out kedua harus HTTP `409`.
9. `GET /api/v1/attendance/today` harus menampilkan `COMPLETED`.
10. `GET /api/v1/attendance/history` harus menampilkan record user tersebut.
11. Token `ADMIN` harus ditolak HTTP `403`.
12. Token `USER` lain tidak dapat melihat record user tersebut.

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


