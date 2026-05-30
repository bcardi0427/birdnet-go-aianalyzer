# Set up environment variables for MSYS2 GCC and TensorFlow Lite C DLL
$scriptDir = $PSScriptRoot
if (-not $scriptDir) { $scriptDir = $PWD.Path }

if (-not $env:TENSORFLOW_PATH) {
    Write-Error "TENSORFLOW_PATH environment variable is not set. Please set it to your TensorFlow header path (see .env.example)."
    exit 1
}

if (-not (Test-Path -Path $env:TENSORFLOW_PATH)) {
    Write-Error "TENSORFLOW_PATH '$env:TENSORFLOW_PATH' does not exist."
    exit 1
}

$env:PATH = "C:\msys64\ucrt64\bin;$scriptDir;" + $env:PATH
$env:CGO_ENABLED = "1"
$env:CGO_CFLAGS = "-I$env:TENSORFLOW_PATH"

$version = "unknown"
if (Test-Path -Path "$scriptDir\version.txt") {
    $version = (Get-Content -Path "$scriptDir\version.txt" -Raw).Trim()
} else {
    $gitVersion = git describe --tags --always 2>$null
    if ($gitVersion) { $version = $gitVersion }
}

$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$ldflags = "-s -w -X 'main.buildDate=$buildDate' -X 'main.version=$version'"

Write-Host "Building birdnet-go.exe with version: $version..."
go build -trimpath -ldflags $ldflags -o birdnet-go.exe .
if ($LASTEXITCODE -eq 0) {
    Write-Host "Build succeeded!" -ForegroundColor Green
} else {
    Write-Host "Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
}
