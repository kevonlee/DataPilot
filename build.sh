#!/bin/bash

echo "========================================"
echo " Database Manager - Build Script"
echo "========================================"
echo ""

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Error: Node.js is not installed"
    echo "Please install Node.js from https://nodejs.org/"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go from https://go.dev/"
    exit 1
fi

echo "[1/4] Installing frontend dependencies..."
cd web
npm install
if [ $? -ne 0 ]; then
    echo "Error: npm install failed"
    exit 1
fi

echo ""
echo "[2/4] Building frontend..."
npm run build
if [ $? -ne 0 ]; then
    echo "Error: Frontend build failed"
    exit 1
fi

cd ..

echo ""
echo "[3/4] Installing Go dependencies..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "Error: go mod tidy failed"
    exit 1
fi

echo ""
echo "[4/4] Building Go binary..."
go build -o dbmanager .
if [ $? -ne 0 ]; then
    echo "Error: Go build failed"
    exit 1
fi

echo ""
echo "========================================"
echo " Build successful!"
echo " Run: ./dbmanager"
echo " Default login: admin / admin"
echo " URL: http://localhost:9090"
echo "========================================"
