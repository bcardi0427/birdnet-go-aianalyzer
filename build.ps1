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

Write-Host "Building birdnet-go.exe..."
go build -trimpath -o birdnet-go.exe .
if ($LASTEXITCODE -eq 0) {
    Write-Host "Build succeeded!" -ForegroundColor Green
} else {
    Write-Host "Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
}
