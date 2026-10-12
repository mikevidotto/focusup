#!/usr/bin/env bash
# Builds the gym server (cmd/gymserver) with its phone page embedded.
#
#   scripts/build-gym.sh            # Raspberry Pi, 64-bit OS (arm64)
#   GOARCH=arm GOARM=7 scripts/build-gym.sh   # 32-bit Pi OS
#
# Output: build/gymserver
set -euo pipefail
cd "$(dirname "$0")/.."

(cd frontend && npm run build:gym)
GOOS="${GOOS:-linux}" GOARCH="${GOARCH:-arm64}" CGO_ENABLED=0 \
    go build -o build/gymserver ./cmd/gymserver

echo "built build/gymserver for ${GOOS:-linux}/${GOARCH:-arm64}"
