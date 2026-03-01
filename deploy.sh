#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$SCRIPT_DIR/../money-manage"
BACKEND_DIR="$SCRIPT_DIR"

echo "=== 1. Build frontend ==="
cd "$FRONTEND_DIR"
npm ci
npm run build

echo "=== 2. Copy dist to backend ==="
rm -rf "$BACKEND_DIR/dist"
cp -r "$FRONTEND_DIR/dist" "$BACKEND_DIR/dist"

echo "=== 3. Start services ==="
cd "$BACKEND_DIR"
docker compose up -d --build

echo ""
echo "=== Done! ==="
echo "App running at http://localhost:8080"
