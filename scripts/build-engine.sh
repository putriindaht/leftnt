#!/bin/sh
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$project_dir/server"
mkdir -p ../client/public/engine
GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -trimpath -o ../client/public/engine/game.wasm ./cmd/ebitengine
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../client/public/engine/wasm_exec.js
echo "Built the browser game in client/public/engine."
