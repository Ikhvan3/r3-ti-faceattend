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

## Admin Office Location Management

Endpoint lokasi kantor hanya untuk access token role `ADMIN`.

### Daftar Lokasi Kantor

- `GET /api/v1/admin/office-locations`

Query: `page` default `1`, `page_size` default `10` maksimum `100`,
`search` pada `name` dan `address`, dan `status` opsional `ACTIVE` atau
`INACTIVE`. Daftar kosong tetap HTTP `200` dengan `items: []`.

### Tambah Lokasi Kantor

- `POST /api/v1/admin/office-locations`

```json
{
  "name": "Kantor Regional 3",
  "address": "Alamat development",
  "latitude": -6.123456,
  "longitude": 106.123456,
  "radius_meters": 100
}
```

Lokasi baru selalu aktif. Field `is_active`, `created_at`, dan `updated_at`
tidak diterima dari body. Latitude harus `-90` sampai `90`, longitude harus
`-180` sampai `180`, dan radius geofence harus `10` sampai `2000` meter.

### Detail dan Update Lokasi Kantor

- `GET /api/v1/admin/office-locations/{id}`
- `PUT /api/v1/admin/office-locations/{id}`

Update hanya menerima `name`, `address`, `latitude`, `longitude`, dan
`radius_meters`. Status hanya diubah melalui endpoint status.

### Status Lokasi Kantor

- `PATCH /api/v1/admin/office-locations/{id}/status`

```json
{
  "is_active": false
}
```

Lokasi yang memiliki assignment lokasi `CURRENT` atau `UPCOMING` berdasarkan
`BUSINESS_TIMEZONE` tidak boleh dinonaktifkan dan mengembalikan HTTP `409`.

## Admin Location Assignment

Endpoint assignment lokasi hanya untuk access token role `ADMIN`. Response
memuat profil pegawai aman dan lokasi kantor; tidak memuat `password_hash`,
token, session, atau attendance record.

Tanggal efektif memakai semantik inklusif: `effective_from` dan
`effective_to` sama-sama termasuk periode assignment. `effective_to = null`
berarti tidak memiliki batas akhir. PostgreSQL mencegah overlap periode untuk
user yang sama.

### Daftar Assignment Lokasi

- `GET /api/v1/admin/location-assignments`

Query: `page`, `page_size`, `search` untuk `employee_number`, `name`, atau
`email`, `user_id`, `office_location_id`, dan `status` opsional `CURRENT`,
`UPCOMING`, atau `ENDED`.

### Tambah Assignment Lokasi

- `POST /api/v1/admin/location-assignments`

```json
{
  "user_id": "00000000-0000-4000-8000-000000000001",
  "office_location_id": "00000000-0000-4000-8000-000000000030",
  "effective_from": "2026-08-06",
  "effective_to": null
}
```

`user_id` wajib dan tidak dapat diganti dengan `employee_number`. User harus
tersedia dengan role `USER`, lokasi harus tersedia dan aktif, dan periode
tidak boleh overlap dengan assignment lokasi user yang sama. Overlap
mengembalikan HTTP `409`.

### Detail dan Akhiri Assignment Lokasi

- `GET /api/v1/admin/location-assignments/{id}`
- `PATCH /api/v1/admin/location-assignments/{id}/end`

```json
{
  "effective_to": "2026-08-31"
}
```

Tanggal akhir bersifat inklusif. Tidak ada endpoint `DELETE`.

## Basic Attendance User

Endpoint attendance hanya untuk access token role `USER`. Backend selalu
mengambil `user_id` dari token dan waktu dari server dengan timezone bisnis
`Asia/Jakarta` secara default. Client tidak mengirim `user_id`, tanggal, atau
waktu absensi. Client juga tidak mengirim `location_id`, jarak, role, atau
timestamp perangkat.

### Attendance Today

Endpoint:

- `GET /api/v1/attendance/today`

Response `data` memuat `attendance_date`, `schedule`, `check_in_at`,
`check_out_at`, `state`, `can_check_in`, `can_check_out`,
`check_in_location`, dan `check_out_location`. Nilai `state` adalah
`NOT_CHECKED_IN`, `CHECKED_IN`, atau `COMPLETED`.

Evidence lokasi pada response aman untuk client:

```json
{
  "office_location_id": "00000000-0000-4000-8000-000000000030",
  "office_location_name": "Kantor Regional 3",
  "accuracy_meters": 12.5,
  "distance_meters": 36.7,
  "inside_geofence": true
}
```

### Check-in

Endpoint:

- `POST /api/v1/attendance/check-in`

Request:

```json
{
  "latitude": -6.98946,
  "longitude": 110.416735,
  "accuracy_meters": 12.5,
  "verification_grant": "opaque-token"
}
```

Check-in berhasil mengembalikan HTTP `201`. Backend mengambil assignment lokasi
aktif berdasarkan tanggal bisnis, menghitung jarak geofence server-side, dan
menyimpan evidence lokasi pada attendance record dalam transaksi yang sama.
Grant wajib terikat ke user dari token, purpose `CHECK_IN`, belum expired, dan
belum consumed. Check-in ganda pada tanggal kerja yang sama mengembalikan HTTP
`409`.

### Check-out

Endpoint:

- `POST /api/v1/attendance/check-out`

Request sama dengan check-in. Check-out berhasil mengembalikan HTTP `200` dan
menyimpan evidence lokasi check-out. Grant wajib terikat ke purpose
`CHECK_OUT`. Check-out sebelum check-in atau check-out ganda mengembalikan HTTP
`409`.

Error geofence:

- HTTP `400` untuk JSON/koordinat invalid.
- HTTP `404` jika assignment lokasi hari ini tidak tersedia.
- HTTP `403` jika lokasi kantor nonaktif atau posisi berada di luar radius.
- HTTP `422` jika `accuracy_meters` melebihi `GEOFENCE_MAX_ACCURACY_METERS`.

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
dan mengulang request attendance satu kali. Mobile hanya mengirim latitude,
longitude, dan accuracy saat check-in/check-out.

### Attendance Location Requirement

Endpoint:

- `GET /api/v1/attendance/location-requirement`

Endpoint ini hanya untuk access token role `USER`. Backend mengambil `user_id`
dari token dan menggunakan `BUSINESS_TIMEZONE` untuk mencari assignment lokasi
hari ini. Request tidak menerima `user_id` dari query atau body.

Response `data` memuat assignment lokasi aktif dan lokasi kantor aktif untuk
pegawai tersebut. Jika assignment lokasi tidak tersedia, backend mengembalikan
HTTP `404` dengan pesan aman. Endpoint ini hanya memberi kebutuhan lokasi;
enforcement geofence tetap dilakukan oleh endpoint check-in/check-out.

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

## Face Enrollment Foundation

Endpoint face enrollment berada di backend Golang dan hanya menerima access
token role `USER`.

Endpoint:

- `GET /api/v1/face/status`
- `POST /api/v1/face/enroll`
- `DELETE /api/v1/face/enrollment`

`GET /api/v1/face/status` mengambil `user_id` dari token. Response untuk user
yang belum enrollment:

```json
{
  "status": "ok",
  "data": {
    "enrolled": false,
    "face_status": "NOT_ENROLLED"
  }
}
```

Response untuk user yang sudah enrollment memuat metadata model dan waktu
enrollment, tetapi tidak pernah memuat embedding:

```json
{
  "status": "ok",
  "data": {
    "enrolled": true,
    "face_status": "ENROLLED",
    "embedding_model": "nama-model",
    "embedding_version": "versi-model",
    "enrolled_at": "2026-08-07T01:00:00Z"
  }
}
```

`POST /api/v1/face/enroll` menyimpan embedding untuk user dari token. Request
tidak menerima `user_id` dari body.

```json
{
  "embedding": [0.1, 0.2, 0.3],
  "embedding_model": "nama-model",
  "embedding_version": "versi-model"
}
```

Backend memvalidasi bahwa akun user masih `ACTIVE`, embedding tidak kosong,
semua nilai finite, model/version didukung registry backend, dan dimensi
embedding sama dengan kontrak model. Duplikasi enrollment sebelum reset
menghasilkan HTTP `409`.

`DELETE /api/v1/face/enrollment` menghapus enrollment milik user dari token.
Setelah reset, `GET /api/v1/face/status` kembali menampilkan
`NOT_ENROLLED` dan user dapat enrollment ulang.

`POST /api/v1/face/verify` menjalankan verifikasi wajah standalone untuk user
dari access token. Endpoint ini tidak menerima `user_id`, `employee_number`,
`threshold`, `similarity`, `verified`, atau enrolled embedding dari client.

Request:

```json
{
  "embedding": [0.1, 0.2, 0.3],
  "embedding_model": "facenet",
  "embedding_version": "shubham0204-facenet-2020-fp32"
}
```

Response sukses:

```json
{
  "status": "ok",
  "data": {
    "verified": true
  }
}
```

Jika wajah kandidat tidak cocok, response tetap HTTP `200` dengan
`"verified": false` karena proses verifikasi berhasil dijalankan. User yang
belum enrollment menghasilkan HTTP `409`. Request malformed, unknown field,
model/version salah, dimensi salah, nilai `NaN`/`Inf`, atau zero vector ditolak
sebagai HTTP `400`.

Backend mengambil enrolled embedding dari `face_profiles` berdasarkan user JWT,
memastikan akun masih `ACTIVE`, memastikan profile berstatus `ENROLLED`, lalu
membandingkan candidate embedding dengan stored embedding menggunakan cosine
similarity. Threshold verifikasi berada di backend melalui
`FACE_VERIFICATION_THRESHOLD`; Flutter tidak mengetahui dan tidak mengirim
threshold. Karena belum ada hasil kalibrasi lokal yang cukup di repository,
nilai environment ini wajib diisi untuk development dan harus dikalibrasi dari
pengujian user yang sudah enrollment sebelum dianggap final.

Catatan tahap ini:

- Backend menerima model embedding produksi berikut:
  - `embedding_model`: `facenet`
  - `embedding_version`: `shubham0204-facenet-2020-fp32`
  - dimensi embedding: `128`
  - file Flutter: `assets/models/facenet.tflite`
- Model lain, versi lain, atau dimensi selain `128` ditolak.
- Embedding tidak boleh dicetak ke log, tidak dikembalikan pada response
  status/enroll/reset/verify, dan tidak dimasukkan ke pesan error.
- Endpoint `/face/verify` tetap standalone dan tidak menerbitkan grant
  attendance.

### Verify For Attendance

Endpoint:

- `POST /api/v1/face/verify-for-attendance`

Endpoint ini hanya menerima access token role `USER` aktif. Request tidak
menerima `user_id`, `employee_id`, `threshold`, `similarity`, `verified`, atau
`expires_at` dari client.

Request:

```json
{
  "purpose": "CHECK_IN",
  "embedding": [0.1, 0.2, 0.3],
  "embedding_model": "facenet",
  "embedding_version": "shubham0204-facenet-2020-fp32"
}
```

`purpose` hanya boleh `CHECK_IN` atau `CHECK_OUT`. Backend memakai pipeline
verifikasi wajah yang sama dengan `/face/verify`: profile harus enrolled,
model dan version harus cocok dengan registry, dimensi embedding harus sesuai,
nilai harus finite dan bukan zero vector, dan threshold tetap dari backend.

Jika wajah cocok, backend membuat verification grant server-side yang singkat,
sekali pakai, dan terikat ke user serta purpose. Raw grant dikirim satu kali ke
Flutter, sedangkan database hanya menyimpan hash token.

Response:

```json
{
  "status": "ok",
  "data": {
    "verification_grant": "opaque-token",
    "expires_at": "2026-08-07T01:02:00Z"
  }
}
```

Response tidak memuat embedding, similarity score, threshold, atau stored face
template. Jika wajah tidak cocok, grant tidak dibuat.

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
