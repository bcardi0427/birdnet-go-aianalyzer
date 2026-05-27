@echo off
REM Sourcing this script sets up the environment variables for MSYS2 GCC and TensorFlow Lite C DLL

if "%TENSORFLOW_PATH%"=="" (
    echo [ERROR] TENSORFLOW_PATH environment variable is not set.
    echo Please set TENSORFLOW_PATH to the path of your TensorFlow directory (see .env.example).
    exit /b 1
)

if not exist "%TENSORFLOW_PATH%" (
    echo [ERROR] TENSORFLOW_PATH directory "%TENSORFLOW_PATH%" does not exist.
    exit /b 1
)

set "PATH=C:\msys64\ucrt64\bin;%~dp0;%PATH%"
set "CGO_ENABLED=1"
set "CGO_CFLAGS=-I%TENSORFLOW_PATH%"
echo Environment configured for MSYS2 UCRT64 GCC and TensorFlow Lite headers.
