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

## Admin Web

```powershell
cd admin-web
npm install
npm run lint
npx tsc --noEmit
npm run dev
```

Konfigurasi awal dicatat di `admin-web/.env.example`.

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
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_SSLMODE`

PowerShell dapat memuat nilai untuk sesi terminal seperti ini:

```powershell
$env:APP_ENV="local"
$env:APP_PORT="8080"
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_NAME="r3_ti_faceattend"
$env:DB_USER="postgres"
$env:DB_PASSWORD=
$env:DB_SSLMODE="disable"
go run ./cmd/api
```

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
