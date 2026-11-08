#!/bin/sh
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$project_dir"
mkdir -p .local/voice-test
./client/node_modules/.bin/tsc --target ES2022 --lib ES2022,DOM --module ES2022 --moduleResolution bundler --skipLibCheck --outDir .local/voice-test client/src/lib/voice-call.ts
cp client/tests/voice-call.html .local/voice-test/index.html
voice_port=${VOICE_TEST_PORT:-8789}
echo "Open http://localhost:$voice_port and choose Run synthetic call test."
echo "This test uses generated tones, without a physical microphone."
cd server
PORT="$voice_port" STATIC_DIR="$project_dir/.local/voice-test" WEBRTC_ICE_SERVERS='[]' exec go run ./cmd/api
