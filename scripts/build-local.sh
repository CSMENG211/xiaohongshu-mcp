#!/bin/sh
set -eu

REPO_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$REPO_DIR"
VERSION="v2.7.4-rednote-fork+$(git describe --always --dirty)"
mkdir -p bin
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o bin/xiaohongshu-mcp .
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/xiaohongshu-login ./cmd/login
