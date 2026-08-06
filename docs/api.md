# API

Dokumen ini mencatat endpoint backend Golang untuk R3 TI FaceAttend.

Base path API adalah `/api/v1`. Jangan menulis token, secret, atau password
nyata ke dokumentasi, commit, issue, atau log.

## Health

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/api/v1/health
```

## Login

Endpoint:

- `POST /api/v1/auth/login`

Request:

```json
{
  "email": "admin.local@example.test",
  "password": "password-development"
}
```

PowerShell:

```powershell
$LoginBody = @{
  email = "admin.local@example.test"
  password = "password-development"
} | ConvertTo-Json

$Login = Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/auth/login" `
  -ContentType "application/json" `
  -Body $LoginBody
```

Response berisi `access_token`, `refresh_token`, `token_type`, `expires_in`,
dan profil user aman. Response tidak memuat `password_hash`.

## Auth Me

Endpoint:

- `GET /api/v1/auth/me`

PowerShell:

```powershell
$Headers = @{
  Authorization = "Bearer $($Login.data.access_token)"
}

Invoke-RestMethod `
  -Method Get `
  -Uri "http://localhost:8080/api/v1/auth/me" `
  -Headers $Headers
```

Catatan mobile:

- Flutter mobile memakai `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`,
  `POST /api/v1/auth/logout`, dan `GET /api/v1/auth/me`.
- Mobile hanya menerima user `USER` dengan status `ACTIVE`.
- Akun `ADMIN`, `INACTIVE`, atau `SUSPENDED` harus ditolak oleh aplikasi
  mobile meskipun response backend valid.

## Admin Ping

Endpoint pengujian role ADMIN:

- `GET /api/v1/admin/ping`

PowerShell:

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "http://localhost:8080/api/v1/admin/ping" `
  -Headers $Headers
```

Endpoint ini membutuhkan access token valid dan role `ADMIN`.

## Employee Management Admin

Endpoint employee hanya untuk pegawai Divisi Teknologi Informasi pada scope
development saat ini. Data employee disimpan sebagai user dengan `role = USER`.
Endpoint ini membutuhkan access token valid dan role `ADMIN`; akun `ADMIN`
tidak dapat dibaca atau diubah melalui endpoint employee.

Response employee tidak pernah memuat `password_hash`.

### Daftar Pegawai

Endpoint:

- `GET /api/v1/admin/employees`

Query parameter:

- `page`, default `1`
- `page_size`, default `10`, maksimum `100`
- `search`, opsional, mencari `employee_number`, `name`, `email`, dan
  `position`
- `status`, opsional: `ACTIVE`, `INACTIVE`, atau `SUSPENDED`

Response `data`:

```json
{
  "items": [],
  "page": 1,
  "page_size": 10,
  "total_items": 0,
  "total_pages": 0
}
```

PowerShell:

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "http://localhost:8080/api/v1/admin/employees?page=1&page_size=10&search=dummy" `
  -Headers $Headers
```

### Tambah Pegawai

Endpoint:

- `POST /api/v1/admin/employees`

Request:

```json
{
  "employee_number": "EMP-DUMMY-001",
  "name": "Pegawai Dummy TI",
  "email": "pegawai.dummy.ti@example.test",
  "initial_password": "password-dummy",
  "phone": null,
  "position": "Staf TI"
}
```

`role` selalu dipaksa menjadi `USER`, `account_status` awal selalu `ACTIVE`,
dan `initial_password` di-hash sebelum disimpan.

PowerShell:

```powershell
$EmployeeBody = @{
  employee_number = "EMP-DUMMY-001"
  name = "Pegawai Dummy TI"
  email = "pegawai.dummy.ti@example.test"
  initial_password = "password-dummy"
  phone = $null
  position = "Staf TI"
} | ConvertTo-Json

$Employee = Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/admin/employees" `
  -Headers $Headers `
  -ContentType "application/json" `
  -Body $EmployeeBody
```

Duplicate `email` atau `employee_number` mengembalikan HTTP `409`.

### Detail Pegawai

Endpoint:

- `GET /api/v1/admin/employees/{id}`

PowerShell:

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "http://localhost:8080/api/v1/admin/employees/$($Employee.data.id)" `
  -Headers $Headers
```

UUID tidak valid mengembalikan HTTP `400`. Pegawai tidak ditemukan
mengembalikan HTTP `404`.

### Update Pegawai

Endpoint:

- `PUT /api/v1/admin/employees/{id}`

Request:

```json
{
  "employee_number": "EMP-DUMMY-001",
  "name": "Pegawai Dummy TI Updated",
  "email": "pegawai.dummy.updated@example.test",
  "phone": "081234567890",
  "position": "Staf TI"
}
```

Field yang dapat diubah hanya `employee_number`, `name`, `email`, `phone`, dan
`position`.

PowerShell:

```powershell
$UpdateBody = @{
  employee_number = "EMP-DUMMY-001"
  name = "Pegawai Dummy TI Updated"
  email = "pegawai.dummy.updated@example.test"
  phone = "081234567890"
  position = "Staf TI"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Put `
  -Uri "http://localhost:8080/api/v1/admin/employees/$($Employee.data.id)" `
  -Headers $Headers `
  -ContentType "application/json" `
  -Body $UpdateBody
```

### Update Status Pegawai

Endpoint:

- `PATCH /api/v1/admin/employees/{id}/status`

Request:

```json
{
  "account_status": "INACTIVE"
}
```

PowerShell:

```powershell
$StatusBody = @{
  account_status = "INACTIVE"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Patch `
  -Uri "http://localhost:8080/api/v1/admin/employees/$($Employee.data.id)/status" `
  -Headers $Headers `
  -ContentType "application/json" `
  -Body $StatusBody
```

Status yang diterima hanya `ACTIVE`, `INACTIVE`, dan `SUSPENDED`.

## Admin Work Schedule Management

Endpoint jadwal kerja hanya untuk access token role `ADMIN`.

### Daftar Jadwal Kerja

- `GET /api/v1/admin/work-schedules`

Query: `page` default `1`, `page_size` default `10` maksimum `100`,
`search` pada `name`, dan `status` opsional `ACTIVE` atau `INACTIVE`.
Daftar kosong tetap HTTP `200` dengan `items: []`.

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "http://localhost:8080/api/v1/admin/work-schedules?page=1&page_size=10&status=ACTIVE" `
  -Headers $Headers
```

### Tambah Jadwal Kerja

- `POST /api/v1/admin/work-schedules`

```json
{
  "name": "Jadwal Kerja Reguler",
  "start_time": "08:00",
  "end_time": "17:00",
  "grace_minutes": 15
}
```

Jadwal baru selalu aktif. Field `is_active`, `created_at`, dan `updated_at`
tidak diterima dari body. Jadwal harus berada dalam hari yang sama,
`end_time` harus setelah `start_time`, dan `grace_minutes` harus `0` sampai
`240` menit. Nama duplikat mengembalikan HTTP `409`.

```powershell
$ScheduleBody = @{
  name = "Jadwal Kerja Reguler"
  start_time = "08:00"
  end_time = "17:00"
  grace_minutes = 15
} | ConvertTo-Json

$Schedule = Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/admin/work-schedules" `
  -Headers $Headers `
  -ContentType "application/json" `
  -Body $ScheduleBody
```

### Detail dan Update Jadwal Kerja

- `GET /api/v1/admin/work-schedules/{id}`
- `PUT /api/v1/admin/work-schedules/{id}`

Update hanya menerima `name`, `start_time`, `end_time`, dan `grace_minutes`.
`is_active` tidak diubah lewat endpoint update.

```powershell
$ScheduleUpdateBody = @{
  name = "Jadwal Kerja Reguler Updated"
  start_time = "08:00"
  end_time = "17:00"
  grace_minutes = 20
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Put `
  -Uri "http://localhost:8080/api/v1/admin/work-schedules/$($Schedule.data.id)" `
  -Headers $Headers `
  -ContentType "application/json" `
  -Body $ScheduleUpdateBody
```

### Status Jadwal Kerja

- `PATCH /api/v1/admin/work-schedules/{id}/status`

```json
{
  "is_active": false
}
```

Jadwal yang memiliki assignment aktif atau masa depan berdasarkan
`BUSINESS_TIMEZONE` tidak boleh dinonaktifkan dan mengembalikan HTTP `409`.
Aktivasi kembali diperbolehkan.

## Admin Schedule Assignment

Endpoint assignment jadwal hanya untuk access token role `ADMIN`. Response
memuat profil pegawai aman dan ringkasan jadwal; tidak memuat
`password_hash`, token, session, atau attendance record.

Tanggal efektif memakai semantik inklusif: `effective_from` dan
`effective_to` sama-sama termasuk periode assignment. `effective_to = null`
berarti tidak memiliki batas akhir. Assignment boleh diakhiri pada tanggal
hari ini; assignment masih berlaku sampai akhir tanggal bisnis tersebut.

### Daftar Assignment

- `GET /api/v1/admin/schedule-assignments`

Query: `page`, `page_size`, `search` untuk `employee_number`, `name`, atau
`email`, `user_id`, `schedule_id`, dan `status` opsional `CURRENT`,
`UPCOMING`, atau `ENDED`. Status dihitung menggunakan `BUSINESS_TIMEZONE`.

```powershell
Invoke-RestMethod `
  -Method Get `
  -Uri "http://localhost:8080/api/v1/admin/schedule-assignments?status=CURRENT" `
  -Headers $Headers
```

### Tambah Assignment

- `POST /api/v1/admin/schedule-assignments`

```json
{
  "user_id": "00000000-0000-4000-8000-000000000001",
  "schedule_id": "00000000-0000-4000-8000-000000000010",
  "effective_from": "2026-08-05",
  "effective_to": null
}
```

`user_id` wajib dan tidak dapat diganti dengan `employee_number`. User harus
tersedia dengan role `USER`, schedule harus tersedia dan aktif, periode tidak
boleh overlap dengan assignment user yang sama. Overlap mengembalikan HTTP
`409`.

```powershell
$AssignmentBody = @{
  user_id = $Employee.data.id
  schedule_id = $Schedule.data.id
  effective_from = "2026-08-05"
  effective_to = $null
} | ConvertTo-Json

$Assignment = Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/api/v1/admin/schedule-assignments" `
  -Headers $Headers `
  -ContentType "application/json" `
  -Body $AssignmentBody
```

### Detail dan Akhiri Assignment

- `GET /api/v1/admin/schedule-assignments/{id}`
- `PATCH /api/v1/admin/schedule-assignments/{id}/end`

```powershell
$EndAssignmentBody = @{
  effective_to = "2026-08-31"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Patch `
  -Uri "http://localhost:8080/api/v1/admin/schedule-assignments/$($Assignment.data.id)/end" `
  -Headers $Headers `
  -ContentType "application/json" `
  -Body $EndAssignmentBody
```

Assignment yang sudah berakhir tetap dapat dibaca. Endpoint ini tidak
menghapus assignment dan tidak mengubah attendance record lama.

## Basic Attendance User

Endpoint attendance hanya untuk access token role `USER`. Backend selalu
mengambil `user_id` dari token dan waktu dari server dengan timezone bisnis
`Asia/Jakarta` secara default. Client tidak mengirim `user_id`, tanggal, atau
waktu absensi.

### Attendance Today

Endpoint:

- `GET /api/v1/attendance/today`

Response `data` memuat `attendance_date`, `schedule`, `check_in_at`,
`check_out_at`, `state`, `can_check_in`, dan `can_check_out`. Nilai `state`
adalah `NOT_CHECKED_IN`, `CHECKED_IN`, atau `COMPLETED`.

### Check-in

Endpoint:

- `POST /api/v1/attendance/check-in`

Body harus kosong. Check-in berhasil mengembalikan HTTP `201`. Check-in ganda
pada tanggal kerja yang sama mengembalikan HTTP `409`.

### Check-out

Endpoint:

- `POST /api/v1/attendance/check-out`

Body harus kosong. Check-out berhasil mengembalikan HTTP `200`. Check-out
sebelum check-in atau check-out ganda mengembalikan HTTP `409`.

### Attendance History

Endpoint:

- `GET /api/v1/attendance/history`

Query parameter:

- `page`, default `1`
- `page_size`, default `10`, maksimum `100`

Response `data` berisi `items`, `page`, `page_size`, `total_items`, dan
`total_pages`. Riwayat hanya memuat data user dari token. Daftar kosong tetap
HTTP `200` dengan `items: []`.

Flutter mobile memanggil endpoint attendance di atas langsung ke Golang API
dengan `API_BASE_URL`. Jika access token kedaluwarsa, mobile melakukan refresh
token satu kali melalui endpoint auth yang sudah ada, menyimpan token rotasi,
dan mengulang request attendance satu kali. Mobile tidak mengirim `user_id`,
tanggal, waktu, timezone, atau role pada request attendance.

## Refresh Token

Endpoint:

- `POST /api/v1/auth/refresh`

PowerShell:

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

Refresh token dirotasi. Token lama tidak boleh dapat dipakai ulang setelah
refresh berhasil. Database hanya menyimpan hash SHA-256 dari refresh token.

## Logout

Endpoint:

- `POST /api/v1/auth/logout`

PowerShell:

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

Logout mencabut session refresh token dan tidak membocorkan apakah token pernah
ada.

## Catatan Next.js

Admin web menggunakan Next.js Route Handler sebagai BFF untuk autentikasi.
Browser admin memanggil endpoint Next.js berikut:

- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `GET /api/auth/me`

Next.js kemudian memanggil Golang REST API di bawah `/api/v1`. Token dari
Golang tidak dikembalikan ke browser sebagai JSON; Next.js menyimpannya dalam
HttpOnly cookie:

- `r3_access_token`
- `r3_refresh_token`

Access token dipakai Next.js server untuk memanggil `/api/v1/auth/me`. Jika
access token expired dan refresh token tersedia, Next.js melakukan refresh
server-side satu kali melalui `/api/v1/auth/refresh`, menyimpan token baru, dan
mengulang validasi profil. Authorization final tetap dilakukan oleh Golang dan
pemeriksaan role `ADMIN` di server Next.js.

Untuk employee management, Server Components admin web membaca list dan detail
langsung dari Golang REST API menggunakan token dari HttpOnly cookie server-side.
Mutasi dari browser tetap melewati Next.js Route Handler/BFF:

- `POST /api/admin/employees`
- `PUT /api/admin/employees/{id}`
- `PATCH /api/admin/employees/{id}/status`

Route Handler tersebut meneruskan request ke Golang endpoint admin employee,
mendukung refresh token satu kali bila access token kedaluwarsa, dan tidak
mengembalikan token ke browser.

Jangan menyimpan token di `localStorage` atau `sessionStorage`.
