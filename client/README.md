# LEFTN'T browser client

Explorer + Navigator multiplayer is at `/multiplayer/`, using Ebitengine Wasm.
Build assets with `../scripts/build-engine.sh`, or run
`docker compose up --build -d` from the repository root. See
[multiplayer](../docs/multiplayer.md) and [Docker](../docs/docker.md).

A Next.js App Router client written in TypeScript. It renders the Go game state
and sends movement requests; collision and win rules remain in the Go domain.

## Run locally

Start the API in one terminal:

```sh
cd server
go run ./cmd/api
```

From the repository root, start the client in another terminal:

```sh
cd client
npm ci
npm run dev
```

Open http://localhost:3000 and select **Start exploring**. Move using arrow keys,
WASD, or the on-screen buttons. Reach G to win; select **New maze** to restart.
**Refresh game** reads the current state from the API without changing it.

The browser sends requests to `/api/games`. Next.js rewrites `/api/:path*` to
`http://127.0.0.1:8080/:path*`, keeping browser requests on the same origin.
To target another backend, set `GAME_API_URL` before starting or building Next.js.
Restart Next.js after changing this setting. This is a server setting, not a
public browser environment variable.

## Structure

- `src/app/page.tsx`: game screen, keyboard controls, and request state.
- `src/app/globals.css`: responsive layout and maze styles.
- `src/lib/game-api.ts`: typed create, read, and move requests.
- `next.config.ts`: API forwarding.

## Checks

```sh
npm run typecheck
npm run build
```

Use Node.js 20.9 or newer supported by the installed Next.js version. Package
versions are recorded in `package-lock.json`. Fonts use Google Fonts with local
sans-serif fallbacks. No external image assets are required.

This is a single-player client with the complete maze visible. Reloading the
page returns to the start screen. Games remain in API memory until it restarts;
games are not persisted to a database or synchronized between browser tabs.
