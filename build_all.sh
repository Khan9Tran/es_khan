#!/usr/bin/env bash
set -e

# Version & Build Metadata
VERSION=${1:-"v2.0.0"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u +%Y-%m-%d)
AUTHOR="Khan9Tran"

LDFLAGS="-s -w \
  -X 'eskhan/internal/version.Version=${VERSION}' \
  -X 'eskhan/internal/version.Commit=${COMMIT}' \
  -X 'eskhan/internal/version.BuildDate=${BUILD_DATE}' \
  -X 'eskhan/internal/version.BuiltBy=${AUTHOR}'"

mkdir -p bin

echo "🚀 Bắt đầu đóng gói ESKhan ${VERSION} (${COMMIT}) - Tác giả: ${AUTHOR}"
echo "--------------------------------------------------------"

echo "📦 1/4 Đang build cho macOS Apple Silicon (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/eskhan-darwin-arm64 ./cmd/eskhan

echo "📦 2/4 Đang build cho macOS Intel (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/eskhan-darwin-amd64 ./cmd/eskhan

echo "📦 3/4 Đang build cho Linux Server / Desktop (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/eskhan-linux-amd64 ./cmd/eskhan

echo "📦 4/4 Đang build cho Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/eskhan-windows-amd64.exe ./cmd/eskhan

echo ""
echo "🎉 Hoàn tất! Danh sách file thực thi (Standalone Binaries):"
ls -lh bin/
