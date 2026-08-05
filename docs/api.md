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
