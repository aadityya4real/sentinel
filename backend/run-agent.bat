@echo off
set AGENT_BIN=W:\sentinel\backend\agent.exe
set API_URL=http://localhost:8080/api/v1/metrics

:: Run agent in background, pipe output to for loop
for /f "usebackq delims=" %%L in ("%AGENT_BIN%") do (
    echo %%L
    curl -s -X POST %API_URL% -H "Content-Type: application/json" -d "%%L" >nul 2>&1
    timeout /t 2 /nobreak >nul
)
