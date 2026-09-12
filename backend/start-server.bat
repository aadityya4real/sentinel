@echo off
set DATABASE_URL=postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable
set REDIS_HOST=localhost
set REDIS_PORT=6379
set APP_PORT=8080
set LOG_LEVEL=info
W:\sentinel\backend\sentinel-server.exe
