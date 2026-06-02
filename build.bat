@echo off
echo ========================================
echo  Database Manager - Build Script
echo ========================================
echo.

:: Check if Node.js is installed
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Node.js is not installed or not in PATH
    echo Please install Node.js from https://nodejs.org/
    pause
    exit /b 1
)

:: Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://go.dev/
    pause
    exit /b 1
)

echo [1/4] Installing frontend dependencies...
cd web
call npm install
if %errorlevel% neq 0 (
    echo Error: npm install failed
    pause
    exit /b 1
)

echo.
echo [2/4] Building frontend...
call npm run build
if %errorlevel% neq 0 (
    echo Error: Frontend build failed
    pause
    exit /b 1
)

cd ..

echo.
echo [3/4] Installing Go dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo Error: go mod tidy failed
    pause
    exit /b 1
)

echo.
echo [4/4] Building Go binary...
go build -o dbmanager.exe .
if %errorlevel% neq 0 (
    echo Error: Go build failed
    pause
    exit /b 1
)

echo.
echo ========================================
echo  Build successful!
echo  Run: dbmanager.exe
echo  Default login: admin / admin
echo  URL: http://localhost:9090
echo ========================================
pause
