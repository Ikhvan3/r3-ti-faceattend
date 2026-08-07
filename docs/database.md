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

## Employee Management

Tahap employee management tidak menambahkan tabel baru. Pegawai Divisi
Teknologi Informasi direpresentasikan oleh baris pada tabel `users` dengan
`role = 'USER'`.

Tidak ada tabel `division`, `office`, `schedule`, lokasi, absensi, geolocation,
face recognition, liveness, payroll, atau laporan pada tahap ini. Scope divisi
Teknologi Informasi tetap bersifat konseptual sesuai batasan prototype.

Endpoint employee admin hanya melakukan operasi pada baris `users` dengan
`role = 'USER'`, sehingga akun `ADMIN` tidak dapat dibaca atau diubah melalui
endpoint tersebut. Password tetap disimpan sebagai hash pada `password_hash`
dan tidak boleh dimunculkan dalam response API.

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

## Basic Attendance

Migration `000003_create_basic_attendance_tables` menambahkan tabel minimum
untuk absensi dasar tanpa GPS, geofence, kamera, face recognition, liveness,
koreksi absensi, cuti, lembur, atau laporan.

### work_schedules

Kolom utama:

- `id UUID PRIMARY KEY`
- `name VARCHAR(100) NOT NULL`
- `start_time TIME NOT NULL`
- `end_time TIME NOT NULL`
- `grace_minutes INTEGER NOT NULL DEFAULT 0`
- `is_active BOOLEAN NOT NULL DEFAULT TRUE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Constraint dan index:

- `work_schedules_name_not_empty`
- `work_schedules_grace_minutes_non_negative`
- `work_schedules_name_unique`
- `work_schedules_active_idx`

### employee_schedule_assignments

Kolom utama:

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `schedule_id UUID NOT NULL REFERENCES work_schedules(id) ON DELETE RESTRICT`
- `effective_from DATE NOT NULL`
- `effective_to DATE NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Constraint dan index:

- `employee_schedule_assignments_date_range_valid`
- `employee_schedule_assignments_user_schedule_from_unique`
- `employee_schedule_assignments_user_period_no_overlap` dari migration
  `000004_prevent_overlapping_schedule_assignments` mencegah satu user memiliki
  assignment jadwal yang periodenya overlap. Constraint memakai `daterange`
  inklusif `[]`; `effective_to NULL` diperlakukan sebagai tanpa batas akhir.
- `employee_schedule_assignments_user_effective_idx`
- `employee_schedule_assignments_schedule_id_idx`

### attendance_records

Kolom utama:

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `schedule_id UUID NOT NULL REFERENCES work_schedules(id) ON DELETE RESTRICT`
- `attendance_date DATE NOT NULL`
- `check_in_at TIMESTAMPTZ NOT NULL`
- `check_out_at TIMESTAMPTZ NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Constraint dan index:

- `attendance_records_check_out_after_check_in`
- `attendance_records_user_date_unique` mencegah lebih dari satu record untuk
  user dan tanggal kerja yang sama.
- `attendance_records_user_date_desc_idx`
- `attendance_records_schedule_id_idx`

Timestamp check-in dan check-out disimpan sebagai `TIMESTAMPTZ`. Tanggal
absensi ditentukan backend menggunakan timezone bisnis, default
`Asia/Jakarta`.

## Migration Pencegahan Overlap Assignment

Migration `000004_prevent_overlapping_schedule_assignments` menambahkan
extension PostgreSQL `btree_gist` bila belum tersedia, lalu menambahkan
exclusion constraint pada `employee_schedule_assignments`.

Semantik periode:

- `effective_from` inklusif.
- `effective_to` inklusif.
- `effective_to NULL` berarti assignment tidak memiliki tanggal akhir.

Contoh ditolak:

- `2026-08-01` sampai `2026-08-31`
- `2026-08-20` sampai `2026-09-10`

Contoh diperbolehkan:

- `2026-08-01` sampai `2026-08-31`
- `2026-09-01` sampai `NULL`

Migration down hanya menghapus constraint overlap. Data assignment tidak
dihapus. Jangan memakai data pegawai asli untuk verifikasi migration atau
contoh query.

## Office Location dan Assignment Lokasi

Migration `000005_create_office_locations_and_assignments` menambahkan fondasi
lokasi kantor untuk geofence. Endpoint check-in/check-out baru wajib menerima
koordinat perangkat dan menyimpan evidence lokasi setelah backend menghitung
jarak server-side.

### office_locations

Kolom utama:

- `id UUID PRIMARY KEY`
- `name VARCHAR(150) NOT NULL`
- `address TEXT NULL`
- `latitude DOUBLE PRECISION NOT NULL`
- `longitude DOUBLE PRECISION NOT NULL`
- `radius_meters INTEGER NOT NULL`
- `is_active BOOLEAN NOT NULL DEFAULT TRUE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Constraint:

- `office_locations_name_not_empty`
- `office_locations_latitude_valid` membatasi latitude `-90` sampai `90`.
- `office_locations_longitude_valid` membatasi longitude `-180` sampai `180`.
- `office_locations_radius_valid` membatasi radius `10` sampai `2000` meter.

### employee_location_assignments

Kolom utama:

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `office_location_id UUID NOT NULL REFERENCES office_locations(id) ON DELETE RESTRICT`
- `effective_from DATE NOT NULL`
- `effective_to DATE NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Semantik periode:

- `effective_from` inklusif.
- `effective_to` inklusif.
- `effective_to NULL` berarti tanpa batas akhir.

Constraint overlap `employee_location_assignments_user_period_no_overlap`
memakai `btree_gist` dan `daterange(..., '[]')` untuk mencegah satu user
memiliki assignment lokasi yang periodenya bertumpang tindih, termasuk pada
request concurrent.

### Bukti Lokasi Attendance

Migration yang sama menambahkan kolom nullable pada `attendance_records`:

- `check_in_location_id`
- `check_in_latitude`
- `check_in_longitude`
- `check_in_accuracy_meters`
- `check_in_distance_meters`
- `check_out_location_id`
- `check_out_latitude`
- `check_out_longitude`
- `check_out_accuracy_meters`
- `check_out_distance_meters`

Semua kolom bukti lokasi nullable agar record lama tetap valid. Untuk record
baru, backend mengisi evidence check-in/check-out dalam transaksi attendance
yang sama. Kolom lokasi memiliki FK ke `office_locations`; kolom koordinat,
accuracy, dan distance memiliki check constraint rentang nilai dasar.

## Face Enrollment Foundation

Migration `000006_create_face_profiles` menambahkan tabel `face_profiles` untuk
menyimpan satu enrollment wajah milik setiap user. Tabel ini hanya menyimpan
embedding numerik dan metadata model; tidak menyimpan gambar wajah, base64
image, frame kamera, atau payload biometrik lain.

### face_profiles

Kolom utama:

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `embedding DOUBLE PRECISION[] NOT NULL`
- `embedding_model VARCHAR(100) NOT NULL`
- `embedding_version VARCHAR(100) NOT NULL`
- `status VARCHAR(20) NOT NULL`
- `enrolled_at TIMESTAMPTZ NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

Constraint:

- `face_profiles_user_unique` menjaga satu user hanya memiliki satu face
  profile aktif.
- `face_profiles_embedding_not_empty` menolak array embedding kosong.
- `face_profiles_embedding_model_not_empty` dan
  `face_profiles_embedding_version_not_empty` menolak metadata model kosong.
- `face_profiles_status_allowed` membatasi status tersimpan ke `ENROLLED`.
- `face_profiles_enrolled_at_required` mewajibkan `enrolled_at` saat status
  `ENROLLED`.

Reset enrollment menghapus row `face_profiles` milik user. Kondisi
`NOT_ENROLLED` direpresentasikan sebagai tidak adanya row, bukan embedding
kosong.

`pgvector` belum menjadi dependency schema tahap ini. Model enrollment mobile
yang dipakai adalah `facenet` version `shubham0204-facenet-2020-fp32` dengan
dimensi `128`, dan dimensi tersebut divalidasi di service berdasarkan registry
model backend. Migration tetap memakai `DOUBLE PRECISION[]` agar tidak mengubah
schema `000006`; jika nanti backend membutuhkan similarity search, migration
lanjutan dapat mengubah storage ke tipe `vector(128)` setelah extension
`pgvector` diverifikasi tersedia.
