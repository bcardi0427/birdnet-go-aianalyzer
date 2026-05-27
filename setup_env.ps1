# Sourcing this script sets up the environment variables for MSYS2 GCC and TensorFlow Lite C DLL
$scriptDir = $PSScriptRoot
if (-not $scriptDir) { $scriptDir = $PWD.Path }

if (-not $env:TENSORFLOW_PATH) {
    Write-Error "TENSORFLOW_PATH environment variable is not set. Please set it to your TensorFlow header path (see .env.example)."
    return
}

if (-not (Test-Path -Path $env:TENSORFLOW_PATH)) {
    Write-Error "TENSORFLOW_PATH '$env:TENSORFLOW_PATH' does not exist."
    return
}

$env:PATH = "C:\msys64\ucrt64\bin;$scriptDir;" + $env:PATH
$env:CGO_ENABLED = "1"
$env:CGO_CFLAGS = "-I$env:TENSORFLOW_PATH"
Write-Host "Environment configured for MSYS2 UCRT64 GCC and TensorFlow Lite headers." -ForegroundColor Green
