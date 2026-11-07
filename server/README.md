# LEFTN'T server

This module provides two separate executables and shared maze domain code.
Run the commands below from this directory.

## Terminal game

```sh
go run ./cmd/play
```

The game generates a 15-column by 11-row maze, starts the player at `(1, 1)`,
and selects a reachable goal using the distance rule in the
[gameplay reference](../docs/gameplay.md).

Use the arrow keys or `W/A/S/D` to move without pressing Enter. Both lowercase
and uppercase letters work. Press `Q` or Ctrl+C to quit. Reaching the goal ends the program. Run it again to
generate a new maze. Stop either executable with Ctrl+C.

## HTTP server

```sh
go run ./cmd/api
```

The server listens on port `8080`, currently fixed in `cmd/api/main.go`.

### GET /health

```sh
curl -i http://localhost:8080/health
```

Returns HTTP `200 OK` and `Content-Type: application/json`:

```json
{"status":"ok","game":"LEFTN'T"}
```

This checks whether the HTTP process responds, not the health of external
services. Unknown routes return `404`; unsupported methods on `/health`
return `405`. HEAD is also accepted by Go's GET route matching.
Gameplay endpoints are `POST /games`, `GET /games/{id}`, and
`POST /games/{id}/moves`. See the [API reference](../docs/api.md) for bodies
and responses. Start the [Next.js client](../client/README.md) to play in a browser.

Interactive input uses `golang.org/x/term` to enable raw terminal mode. The
previous terminal settings are restored when the game returns, including on
Q, Ctrl+C, Ctrl+D, input errors, and a win. Piped input is also supported.

## Validation

```sh
go test -v ./...
go vet ./...
```

To run only goal selection tests:

```sh
go test -v ./internal/domain -run TestRandomGoal
```

See the [architecture reference](../docs/architecture.md) for package
responsibilities. Tests live beside the domain code in `internal/domain`.
