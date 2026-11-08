# Docker and deployment preparation

## Start and stop

Start Docker Desktop or Docker Engine with Compose. From the repository root:

```sh
docker compose up --build -d
```

Open `http://localhost:8080/multiplayer/`. Solo pixel art is at `/engine/`,
the original browser prototype at `/`, and health at `/health`.

```sh
docker compose logs -f game
docker compose down
```

Choose another host port with:

```sh
LEFTNT_PORT=8090 docker compose up --build -d
```

The default listener is local-only (`127.0.0.1`). Two players can test using
separate tabs or browsers on this computer. For devices on the same Wi-Fi, set
`LEFTNT_BIND_ADDRESS=0.0.0.0` in a root `.env` file and run `docker compose up -d`.
Open `http://YOUR_LAPTOP_LAN_IP:8080/multiplayer/` on the other device. Changing
the port binding recreates the container and clears existing rooms. The LAN
address can change when reconnecting to Wi-Fi. Voice microphone access on a LAN
IP requires trusted HTTPS. Public Internet play also needs hosting and HTTPS.

## Inside the image

The multi-stage Dockerfile builds:

1. Ebitengine WebAssembly and the matching Go `wasm_exec.js` runtime.
2. A Go HTTP/WebSocket server without native graphics dependencies.
3. Next.js static pages with the engine assets included.
4. A small non-root runtime with only the binary and browser files.

Go serves pages, API, and sockets on one origin. No Node server is required
inside the final container. Compose uses a read-only filesystem and drops Linux
capabilities. Generated Wasm, local notes, dependencies, and local build output
stay out of Git and the build context.

## Runtime settings

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | Port inside the container |
| `STATIC_DIR` | `/app/public` in the image | Exported browser assets |
| `ALLOWED_ORIGINS` | Local Next.js dev origins | Extra allowed browser origins, comma separated |
| `LEFTNT_PORT` | `8080` | Compose host port |
| `LEFTNT_BIND_ADDRESS` | `127.0.0.1` | Host interface; `0.0.0.0` permits LAN access |
| `WEBRTC_ICE_SERVERS` | `[]` | Runtime STUN/TURN server JSON for voice |

Same-origin sockets are allowed automatically. For a future split-origin
hosting setup, configure the exact frontend origin including scheme and port.
The launcher selects `wss:` automatically when opened through HTTPS.

See [voice chat](voice.md) for microphone behavior and ICE configuration. Root
`.env` files are excluded from Git and the Docker build context.

## Public hosting next

This version uses one in-memory game server. Use one replica; rooms disappear
on restart or redeploy. Multiple replicas will need shared state/room ownership
and cross-instance broadcasting.

Choose a host supporting long-lived WebSockets. Put HTTPS in front of the
container, preserve socket upgrades, and route the page/API to the same public
host. No public deployment or hosting account is configured yet.

The server handles termination signals, closes room streams, and gives HTTP
requests a bounded shutdown period. `/health` is available for the hosting
platform's HTTP health check.
