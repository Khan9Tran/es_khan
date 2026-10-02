#!/usr/bin/env bash
set -e

mkdir -p bin

echo "📦 1/4 Đang build cho macOS Apple Silicon (M1/M2/M3/M4)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/eskhan-darwin-arm64 ./cmd/eskhan

echo "📦 2/4 Đang build cho macOS Intel (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/eskhan-darwin-amd64 ./cmd/eskhan

echo "📦 3/4 Đang build cho Linux Server / Desktop (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/eskhan-linux-amd64 ./cmd/eskhan

echo "📦 4/4 Đang build cho Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/eskhan-windows-amd64.exe ./cmd/eskhan

echo ""
echo "🎉 Hoàn tất! Danh sách file thực thi độc lập (Standalone Binaries):"
ls -lh bin/
