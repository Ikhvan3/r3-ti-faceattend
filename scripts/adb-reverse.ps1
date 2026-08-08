param(
    [int]$BackendPort = 8080
)

$ErrorActionPreference = "Stop"

function Get-AdbPath {
    $candidates = @()
    if ($env:ANDROID_HOME) {
        $candidates += Join-Path $env:ANDROID_HOME "platform-tools\adb.exe"
    }
    if ($env:ANDROID_SDK_ROOT) {
        $candidates += Join-Path $env:ANDROID_SDK_ROOT "platform-tools\adb.exe"
    }
    if ($env:LOCALAPPDATA) {
        $candidates += Join-Path $env:LOCALAPPDATA "Android\Sdk\platform-tools\adb.exe"
    }

    foreach ($candidate in $candidates) {
        if (Test-Path -LiteralPath $candidate) {
            return $candidate
        }
    }

    $command = Get-Command adb -ErrorAction SilentlyContinue
    if ($command) {
        return $command.Source
    }

    throw "adb tidak ditemukan. Pastikan Android SDK platform-tools sudah terpasang dan adb ada di PATH."
}

function Invoke-AdbChecked {
    param(
        [string]$AdbPath,
        [string[]]$Arguments
    )

    $output = & $AdbPath @Arguments 2>&1
    if ($LASTEXITCODE -ne 0) {
        $message = ($output | Out-String).Trim()
        if ([string]::IsNullOrWhiteSpace($message)) {
            $message = "exit code $LASTEXITCODE"
        }
        throw "adb gagal: $message"
    }

    return @($output)
}

function Get-PhysicalDeviceId {
    param([string]$AdbPath)

    $rawDevices = Invoke-AdbChecked -AdbPath $AdbPath -Arguments @("devices")
    $devices = @(
        $rawDevices |
            Select-Object -Skip 1 |
            ForEach-Object { $_.ToString().Trim() } |
            Where-Object { $_ -match "\sdevice$" } |
            ForEach-Object { ($_ -split "\s+")[0] } |
            Where-Object { $_ -and $_ -notmatch "^emulator-" }
    )

    if ($devices.Count -eq 0) {
        throw "Tidak ada perangkat Android fisik yang aktif. Sambungkan perangkat dan aktifkan USB debugging."
    }
    if ($devices.Count -gt 1) {
        throw "Lebih dari satu perangkat Android fisik terdeteksi. Sisakan satu perangkat aktif lalu jalankan ulang script."
    }

    return [string]$devices[0]
}

$adb = Get-AdbPath
$deviceId = Get-PhysicalDeviceId -AdbPath $adb

Write-Host "Perangkat fisik terdeteksi: $deviceId"

Invoke-AdbChecked -AdbPath $adb -Arguments @(
    "-s", $deviceId,
    "reverse", "tcp:$BackendPort", "tcp:$BackendPort"
) | Out-Null

$reverseList = Invoke-AdbChecked -AdbPath $adb -Arguments @(
    "-s", $deviceId,
    "reverse", "--list"
)

$reverseText = ($reverseList | Out-String)
if ($reverseText -notmatch "tcp:$BackendPort\s+tcp:$BackendPort") {
    throw "ADB reverse tcp:$BackendPort belum terpasang untuk perangkat $deviceId."
}

$reverseList | ForEach-Object { Write-Host $_ }
Write-Host "ADB reverse aktif untuk $deviceId: http://127.0.0.1:$BackendPort -> komputer lokal:$BackendPort"
