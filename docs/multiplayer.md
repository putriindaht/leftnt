# Explorer + Navigator multiplayer

## Play locally

From the repository root:

```sh
docker compose up --build -d
```

Open `http://localhost:8080/multiplayer/`. Create a room as Explorer, then open
a second tab and join its eight-character code as Navigator. The default
Compose listener is local-only; this invite is not an Internet deployment.

The Explorer moves with arrows/WASD or the game buttons. The Navigator uses
those controls to send directions while seeing the entire map. Click the game
for keyboard focus. Explorer R / New starts a synchronized new maze. T switches
hedge/wood locally. Game buttons also accept touch.

Both players can choose **Join voice** to talk while playing. Voice has separate
mute/leave controls and uses WebRTC. See [voice chat](voice.md) for permissions,
local testing, and STUN/TURN settings for public hosting.

Find the friend in red. The Explorer starts at radius one; two collectible
lanterns grant radius three for 20 seconds, with brighter terrain. Another
pickup refills time. Navigator vision is always complete. Rescue reveals the
map to both players.

## Sessions and reconnects

- Creator receives the Explorer seat; one joiner receives the Navigator seat.
- Each seat has a random token. The room code cannot control an existing seat.
- The token travels in the first socket message, not its network URL.
- Session storage remembers the current tab's seat after refresh. Invite links
  contain only the room code. Do not share your session token.
- One socket is allowed per seat. A duplicate window receives an error.
- Both sockets must be connected for movement, hints, restart, and light time.
  Disconnect pauses the room; the engine retries every second and resumes state.
- Commands queued before disconnect are discarded. Server state decides movement.
- Rooms live in one process. Server restart clears them: leave an expired room
  and create a new one. Unoccupied rooms expire after 20 minutes. Capacity is
  256 rooms per process.

## Architecture

| Location | Responsibility |
| --- | --- |
| `server/internal/domain/room.go` | Roles, movement permission, hints, visibility |
| `server/internal/domain/solo.go` | Lantern placement, pickups, and light duration |
| `server/internal/usecase/room.go` | Sessions, filtered snapshots, subscriptions, ticks |
| `server/internal/storage/memory/room.go` | Atomic room access under a mutex |
| `server/internal/delivery/http/room_handler.go` | HTTP sessions and socket transport |
| `server/internal/protocol/room.go` | Shared wire message types |
| `server/internal/client/room.go` | Background Go socket client with reconnect |
| `server/cmd/ebitengine/multiplayer.go` | Input, interpolation, pixel-art room HUD |
| `client/src/app/multiplayer/page.tsx` | Browser create/join lobby and game iframe |

The server serializes only visible terrain and entities for Explorer. Unknown
cells are `null`, and hidden entity coordinates are not included. Navigator
receives a full view. The client never receives the Explorer's hidden maze.

Updates run atomically inside repository operations. Each seat has a bounded
latest-state channel, so a slow client gets the newest snapshot rather than
an unlimited queue. Socket writes have deadlines and a heartbeat. Every command
is authorized by its seat; browser origins are checked during socket upgrades.

## HTTP and WebSocket protocol

Direct routes and their `/api`-prefixed aliases are available:

- `POST /api/rooms`: create; response `{code, token, role}`.
- `POST /api/rooms/{code}/join`: claim Navigator.
- `GET /api/rooms/{code}/ws`: socket upgrade.

First client message:

```json
{"type":"authenticate","token":"YOUR_SESSION_TOKEN"}
```

Following commands:

```json
{"type":"move","direction":"right"}
{"type":"hint","direction":"down"}
{"type":"restart"}
```

Explorer can move/restart; Navigator can send hints. Each seat may act once
per 180 ms. Invalid directions, unsupported messages, and paused actions are
rejected. Server messages have type `state`, `result`, or `error`. State includes
filtered cells, player, optional friend, visible lanterns, counts, light time,
radius, phase, hint, round, and revision. A move result includes `moved`.

## Development without Docker

From the root, build browser assets:

```sh
./scripts/build-engine.sh
```

Start Go and Next.js in separate terminals:

```sh
cd server
go run ./cmd/api
```

```sh
cd client
npm ci
npm run dev -- --hostname 127.0.0.1
```

Open `http://127.0.0.1:3000/multiplayer/`. The development rewrite forwards
HTTP requests to Go; the game socket connects directly to port 8080. Set
`NEXT_PUBLIC_GAME_WS_URL` to another WebSocket server base URL when needed.
For example, an API on port 8081 uses `ws://127.0.0.1:8081`. The API permits
the default local frontend origins. Production Docker uses one origin for both.
Rebuild the engine after editing its Go code.

For desktop multiplayer, use the same room session from `server`:

```sh
LEFTNT_WS_URL=ws://127.0.0.1:8080/api/rooms/ROOMCODE/ws \
LEFTNT_ROOM_TOKEN=YOUR_SESSION_TOKEN \
EBITENGINE_GRAPHICS_LIBRARY=opengl go run ./cmd/ebitengine
```

Close the browser game for that seat first. Without multiplayer variables,
the engine plays solo locally.

## Validation

```sh
cd server
go test -race ./internal/...
go test ./cmd/...
go vet ./...
```

Tests open real sockets against a test server. Coverage includes two-player
synchronization, role permissions, hidden data, hints, reconnects, tokens, and
origins. Application tests cover light time, disconnect pauses, rescue/reveal,
restart, and room expiry.
