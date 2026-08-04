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

Pada tahap Next.js berikutnya, Next.js akan menerima token dari backend Golang
dan menyimpannya melalui server-side HttpOnly cookie. Tahap ini belum
mengimplementasikan Next.js.
