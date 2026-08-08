$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$envPath = Join-Path $repoRoot "backend\.env"

if (-not (Test-Path $envPath)) {
    throw "backend/.env belum ada. Jalankan scripts/setup-backend-env.ps1 terlebih dahulu."
}

$content = Get-Content -Raw $envPath
if ($content -match "(?m)^FACE_DUPLICATE_ENROLLMENT_THRESHOLD=.+$") {
    Write-Host "FACE_DUPLICATE_ENROLLMENT_THRESHOLD sudah tersedia di backend/.env."
    exit 0
}

Write-Host "Tambahkan konfigurasi duplicate biometric enrollment."
Write-Host "Gunakan nilai development yang akan Anda evaluasi; jangan anggap nilai ini sebagai threshold produksi sebelum kalibrasi FAR/FRR."
$threshold = Read-Host "FACE_DUPLICATE_ENROLLMENT_THRESHOLD"
if ([string]::IsNullOrWhiteSpace($threshold)) {
    throw "FACE_DUPLICATE_ENROLLMENT_THRESHOLD wajib diisi."
}

$parsed = 0.0
if (-not [double]::TryParse(
    $threshold,
    [Globalization.NumberStyles]::Float,
    [Globalization.CultureInfo]::InvariantCulture,
    [ref]$parsed
)) {
    throw "Threshold harus berupa angka desimal dengan titik."
}
if ($parsed -lt -1 -or $parsed -gt 1) {
    throw "Threshold harus berada di antara -1 dan 1."
}

$append = "`r`n# 1:N duplicate biometric enrollment protection.`r`nFACE_DUPLICATE_ENROLLMENT_THRESHOLD=$threshold`r`nFACE_DUPLICATE_SEARCH_TOP_K=20`r`n"
[IO.File]::AppendAllText($envPath, $append, [Text.UTF8Encoding]::new($false))

Write-Host "Konfigurasi duplicate biometric berhasil ditambahkan tanpa mengubah JWT secret atau password database."
