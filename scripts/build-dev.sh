#!/usr/bin/env bash
# Build a local development binary as clockyDEV (does not overwrite release clocky).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="${CLOCKY_DEV_VERSION:-0.0.0-dev}"
OUT="$ROOT/clockyDEV"
LDFLAGS="-X github.com/LuizFer1/Clocky/internal/version.Version=${VERSION}"

echo "Building $OUT (version $VERSION)..."
go build -ldflags "$LDFLAGS" -o "$OUT" ./cmd/clocky
"$OUT" version
echo
echo "Run with: ./clockyDEV <command>"
