param(
    [int]$BackendPort = 8080
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$apiBaseUrl = "http://127.0.0.1:$BackendPort/api/v1"

& (Join-Path $PSScriptRoot "adb-reverse.ps1") -BackendPort $BackendPort

Push-Location (Join-Path $root "mobile")
try {
    flutter run "--dart-define=API_BASE_URL=$apiBaseUrl"
}
finally {
    Pop-Location
}
