param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$backendDir = Join-Path $repoRoot "backend"
$envPath = Join-Path $backendDir ".env"

if ((Test-Path $envPath) -and -not $Force) {
    Write-Host "backend/.env sudah ada. Tidak ada perubahan yang dilakukan."
    Write-Host "Gunakan scripts/upgrade-face-duplicate-config.ps1 untuk menambah konfigurasi duplicate biometric tanpa membuat ulang secret."
    Write-Host "Jalankan dengan -Force hanya jika Anda memang ingin membuat ulang seluruh konfigurasi lokal."
    exit 0
}

function ConvertFrom-SecureStringPlainText {
    param([Security.SecureString]$SecureValue)

    $credential = [PSCredential]::new("local", $SecureValue)
    return $credential.GetNetworkCredential().Password
}

function ConvertTo-DotEnvValue {
    param([string]$Value)

    if ($Value -match "[\r\n]") {
        throw "Nilai .env tidak boleh mengandung baris baru."
    }

    $escaped = $Value.Replace("\", "\\").Replace('"', '\"')
    return '"' + $escaped + '"'
}

function Read-Threshold {
    param(
        [string]$Prompt,
        [string]$Name
    )

    $value = Read-Host $Prompt
    if ([string]::IsNullOrWhiteSpace($value)) {
        throw "$Name wajib diisi dengan nilai yang sudah dievaluasi untuk model face project ini."
    }

    $parsed = 0.0
    if (-not [double]::TryParse(
        $value,
        [Globalization.NumberStyles]::Float,
        [Globalization.CultureInfo]::InvariantCulture,
        [ref]$parsed
    )) {
        throw "$Name harus berupa angka desimal dengan titik."
    }
    if ($parsed -lt -1 -or $parsed -gt 1) {
        throw "$Name harus berada di antara -1 dan 1."
    }

    return $value
}

Write-Host "Setup backend lokal R3 TI FaceAttend"
Write-Host "Secret tidak akan dicetak ke terminal atau disimpan ke Git."
Write-Host ""

$dbHost = Read-Host "DB host [localhost]"
if ([string]::IsNullOrWhiteSpace($dbHost)) { $dbHost = "localhost" }

$dbPort = Read-Host "DB port [5432]"
if ([string]::IsNullOrWhiteSpace($dbPort)) { $dbPort = "5432" }

$dbName = Read-Host "DB name [r3_ti_faceattend]"
if ([string]::IsNullOrWhiteSpace($dbName)) { $dbName = "r3_ti_faceattend" }

$dbUser = Read-Host "DB user [postgres]"
if ([string]::IsNullOrWhiteSpace($dbUser)) { $dbUser = "postgres" }

$dbPasswordSecure = Read-Host "DB password" -AsSecureString
$dbPassword = ConvertFrom-SecureStringPlainText $dbPasswordSecure

$faceThreshold = Read-Threshold `
    "FACE_VERIFICATION_THRESHOLD (1:1, gunakan nilai yang sudah Anda uji)" `
    "FACE_VERIFICATION_THRESHOLD"

$duplicateThreshold = Read-Threshold `
    "FACE_DUPLICATE_ENROLLMENT_THRESHOLD (1:N, nilai development; kalibrasi produksi terpisah)" `
    "FACE_DUPLICATE_ENROLLMENT_THRESHOLD"

$secretBytes = New-Object byte[] 48
$rng = [Security.Cryptography.RandomNumberGenerator]::Create()
try {
    $rng.GetBytes($secretBytes)
} finally {
    $rng.Dispose()
}
$authSecret = [Convert]::ToBase64String($secretBytes)

$content = @"
# Local development only. Never commit this file.
APP_ENV=local
APP_PORT=8080
BUSINESS_TIMEZONE=Asia/Jakarta
DB_HOST=$(ConvertTo-DotEnvValue $dbHost)
DB_PORT=$(ConvertTo-DotEnvValue $dbPort)
DB_NAME=$(ConvertTo-DotEnvValue $dbName)
DB_USER=$(ConvertTo-DotEnvValue $dbUser)
DB_PASSWORD=$(ConvertTo-DotEnvValue $dbPassword)
DB_SSLMODE=disable

AUTH_ACCESS_TOKEN_SECRET=$(ConvertTo-DotEnvValue $authSecret)
AUTH_ACCESS_TOKEN_TTL_MINUTES=15
AUTH_REFRESH_TOKEN_TTL_HOURS=168
AUTH_TOKEN_ISSUER=r3-ti-faceattend-api
AUTH_TOKEN_AUDIENCE=r3-ti-faceattend-client

GEOFENCE_MAX_ACCURACY_METERS=50
FACE_VERIFICATION_THRESHOLD=$faceThreshold
FACE_DUPLICATE_ENROLLMENT_THRESHOLD=$duplicateThreshold
FACE_DUPLICATE_SEARCH_TOP_K=20
FACE_ATTENDANCE_GRANT_TTL_SECONDS=120

# Optional local seed configuration.
ADMIN_EMPLOYEE_NUMBER=
ADMIN_NAME=
ADMIN_EMAIL=
ADMIN_PASSWORD=
ADMIN_POSITION=
DEV_ATTENDANCE_USER_EMAIL=
DEV_ATTENDANCE_SCHEDULE_NAME=Jadwal Kerja Dummy TI
DEV_ATTENDANCE_START_TIME=08:00
DEV_ATTENDANCE_END_TIME=17:00
DEV_ATTENDANCE_GRACE_MINUTES=15
"@

[IO.File]::WriteAllText(
    $envPath,
    $content,
    [Text.UTF8Encoding]::new($false)
)

$dbPassword = $null
$authSecret = $null
[Array]::Clear($secretBytes, 0, $secretBytes.Length)

Write-Host ""
Write-Host "backend/.env berhasil dibuat."
Write-Host "File ini di-ignore Git dan tidak boleh di-commit."
Write-Host ""
Write-Host "Selanjutnya cukup:"
Write-Host "  cd backend"
Write-Host "  go run ./cmd/api"
