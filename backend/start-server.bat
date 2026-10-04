@echo off
setlocal
cd /d "%~dp0"

if not defined SENTINEL_API_TOKEN (
    echo Set SENTINEL_API_TOKEN to a random value of at least 32 characters before starting Sentinel.
    exit /b 1
)

go run ./cmd/server
exit /b %ERRORLEVEL%
