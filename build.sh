#!/usr/bin/env bash
set -euo pipefail

OUT="shelf"

echo "building..."
GOARCH=amd64 GOOS=linux CGO_ENABLED=0 \
  go build -trimpath -ldflags="-s -w" -o "$OUT" .

echo "built: $(du -sh "$OUT" | cut -f1)  $OUT"

if command -v upx &>/dev/null; then
  upx --best --lzma "$OUT"
  echo "packed: $(du -sh "$OUT" | cut -f1)  $OUT"
else
  echo "upx not found, skipping packing"
fi
