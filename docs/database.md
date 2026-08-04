# Database

Dokumen ini mencatat mekanisme migration database lokal untuk R3 TI FaceAttend.

## Mekanisme Migration

Backend menggunakan file SQL migration murni di folder `backend/migrations`.
Setiap migration memiliki pasangan file:

- `*.up.sql` untuk menerapkan perubahan schema.
- `*.down.sql` untuk rollback perubahan schema.

Migration dijalankan dengan CLI `golang-migrate/migrate`. CLI ini dipilih
karena mendukung PostgreSQL, mendukung `up` dan `down`, tidak membutuhkan ORM,
dan tidak menjadi dependency kompilasi backend Golang. Backend tetap dapat
dikompilasi walaupun CLI migration belum terpasang.

## Instalasi CLI

CLI migration tidak dianggap sudah tersedia di mesin lokal. Instal salah satu
cara berikut.

Menggunakan Go:

```powershell
go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Pastikan folder binary Go ada di `PATH`. Pada Windows biasanya:

```powershell
$env:PATH="$env:PATH;$env:USERPROFILE\go\bin"
migrate -version
```

Alternatifnya, unduh binary resmi `migrate` untuk Windows dari rilis GitHub
`golang-migrate/migrate`, lalu letakkan binary tersebut di folder yang masuk
ke `PATH`.

## Konfigurasi URL Database

Jangan menulis password database ke dokumentasi atau commit repository.
Bangun URL database dari environment variable di sesi PowerShell.

```powershell
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_NAME="r3_ti_faceattend"
$env:DB_USER="postgres"
$env:DB_PASSWORD=""
$env:DB_SSLMODE="disable"

$DatabaseUrl = "postgres://$($env:DB_USER):$($env:DB_PASSWORD)@$($env:DB_HOST):$($env:DB_PORT)/$($env:DB_NAME)?sslmode=$($env:DB_SSLMODE)"
```

Jika user PostgreSQL lokal tidak memakai password, nilai `DB_PASSWORD` dapat
dibiarkan kosong sesuai konfigurasi lokal.

## Menjalankan Migration Up

Dari folder `backend`:

```powershell
cd backend
migrate -path migrations -database $DatabaseUrl up
```

## Menjalankan Migration Down

Rollback satu versi terakhir:

```powershell
cd backend
migrate -path migrations -database $DatabaseUrl down 1
```

Rollback semua versi:

```powershell
cd backend
migrate -path migrations -database $DatabaseUrl down -all
```

## Melihat Versi Migration

```powershell
cd backend
migrate -path migrations -database $DatabaseUrl version
```

Jika belum ada migration yang diterapkan, CLI dapat menampilkan status
`Nil version`.

## Verifikasi Melalui psql

Pastikan tabel `users` sudah ada:

```powershell
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "\dt users"
```

Lihat kolom, constraint, dan index:

```powershell
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "\d users"
```

Setelah rollback, pastikan tabel sudah terhapus:

```powershell
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "SELECT to_regclass('public.users');"
```

Nilai hasil `to_regclass` harus kosong atau `NULL` setelah migration down.

## Verifikasi Melalui pgAdmin

1. Buka koneksi PostgreSQL lokal di pgAdmin.
2. Pilih database `r3_ti_faceattend`.
3. Buka `Schemas` > `public` > `Tables`.
4. Setelah migration up, pastikan tabel `users` muncul.
5. Buka tab `Columns`, `Constraints`, dan `Indexes` untuk memeriksa schema.
6. Setelah migration down, refresh node `Tables` dan pastikan `users` tidak ada.

## Schema Awal users

Migration pertama hanya membuat tabel `users`. Tidak ada akun contoh, password,
seed admin, atau tabel lain.

Kolom:

- `id UUID PRIMARY KEY`
- `employee_number VARCHAR(50) NOT NULL`
- `name VARCHAR(150) NOT NULL`
- `email VARCHAR(255) NOT NULL`
- `password_hash TEXT NOT NULL`
- `phone VARCHAR(30)`
- `position VARCHAR(100)`
- `role VARCHAR(10) NOT NULL`
- `account_status VARCHAR(20) NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Constraint dan index:

- `users_employee_number_unique` untuk memastikan nomor karyawan unik.
- `users_email_lower_unique` untuk memastikan email unik tanpa membedakan
  huruf besar dan kecil.
- `users_role_allowed` membatasi role ke `ADMIN` atau `USER`.
- `users_account_status_allowed` membatasi status akun ke `ACTIVE`,
  `INACTIVE`, atau `SUSPENDED`.
- Check constraint `*_not_empty` memastikan nilai wajib tidak hanya string
  kosong atau spasi.

## Seed Admin Lokal

Seed admin dijalankan setelah migration `users` sudah diterapkan. Command
`seed-admin` tidak menjalankan migration otomatis dan tidak membuat tabel baru.

Environment yang dibutuhkan:

- `ADMIN_EMPLOYEE_NUMBER`
- `ADMIN_NAME`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`
- `ADMIN_POSITION`
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_SSLMODE`

Password admin development minimal 8 karakter. Password di-hash dengan bcrypt
sebelum disimpan ke kolom `password_hash`.

Contoh menjalankan melalui PowerShell:

```powershell
cd backend
$env:ADMIN_EMPLOYEE_NUMBER="ADMIN-LOCAL"
$env:ADMIN_NAME="Admin Lokal"
$env:ADMIN_EMAIL="admin.local@example.test"
$env:ADMIN_PASSWORD="minimal-8-karakter"
$env:ADMIN_POSITION="Administrator TI"
go run ./cmd/seed-admin
go run ./cmd/seed-admin
Remove-Item Env:ADMIN_PASSWORD
```

Verifikasi admin tanpa menampilkan hash penuh:

```powershell
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "SELECT employee_number, name, email, role, account_status, created_at, password_hash <> '' AS password_hash_filled FROM users WHERE role = 'ADMIN';"
```

Verifikasi idempotensi:

```powershell
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "SELECT COUNT(*) AS admin_count FROM users WHERE role = 'ADMIN' AND lower(email) = lower('admin.local@example.test');"
```

Pastikan password tidak tersimpan plaintext:

```powershell
psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "SELECT password_hash = 'minimal-8-karakter' AS password_is_plaintext FROM users WHERE lower(email) = lower('admin.local@example.test');"
```

## Migration auth_sessions

Migration `000002_create_auth_sessions_table` menambahkan tabel
`auth_sessions` untuk menyimpan session refresh token.

Kolom:

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `refresh_token_hash TEXT NOT NULL`
- `expires_at TIMESTAMPTZ NOT NULL`
- `revoked_at TIMESTAMPTZ NULL`
- `last_used_at TIMESTAMPTZ NULL`
- `created_ip INET NULL`
- `user_agent TEXT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Index:

- `auth_sessions_refresh_token_hash_unique` untuk lookup refresh token hash.
- `auth_sessions_user_id_idx` untuk lookup session per user.
- `auth_sessions_active_expires_at_idx` untuk session aktif yang belum revoked.

Refresh token plaintext tidak disimpan di database. Backend hanya menyimpan
hash SHA-256.
