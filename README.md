# LEFTN'T 🧭

**Two players. One maze. Zero sense of direction.**

LEFTN'T is a cooperative maze game concept: an Explorer moves with limited
visibility while a Navigator sees the map and gives directions.

The current implementation is a single-player terminal prototype and a separate
HTTP server with a health endpoint. Multiplayer, the browser client, fog of war,
and the countdown timer are not implemented yet.

## Current features

- Random maze generation with connected paths and closed outer borders.
- Movement with wall and boundary checks.
- A random goal at least 70% of the farthest reachable distance from the start.
- A win message when the player reaches the goal.
- A JSON health endpoint.
- Automated tests for movement, generation, distances, and goal selection.

## Requirements

Go 1.27.3 or newer, as specified in `server/go.mod`, and Git to clone the project.
No external services are required. Go downloads the terminal support dependency
(`golang.org/x/term`) automatically on the first run.

## Play in the terminal

```sh
git clone git@github.com:putriindaht/leftnt.git
cd leftnt/server
go run ./cmd/play
```

Use the arrow keys or `W/A/S/D` to move immediately, without Enter.
Press `Q` or Ctrl+C to quit.
Reach `G` to win. `P` is the player, `.` is a path, and `#` is a wall.

## Run the HTTP server

From the `server` directory:

```sh
go run ./cmd/api
```

In another terminal:

```sh
curl -i http://localhost:8080/health
```

Expected: HTTP `200 OK`, content type `application/json`, and body:

```json
{"status":"ok","game":"LEFTN'T"}
```

The terminal game runs locally and does not connect to the HTTP server.

## Run checks

From the `server` directory:

```sh
go test -v ./...
go vet ./...
```

## Repository structure

```text
leftnt/
├── docs/                   # Architecture and gameplay reference
├── server/
│   ├── cmd/api/            # HTTP server entry point
│   ├── cmd/play/           # Terminal game entry point
│   ├── internal/
│   │   ├── delivery/http/  # HTTP routes and handlers
│   │   ├── domain/         # Maze, player, goal rules, and tests
│   │   └── usecase/        # Application operations
│   ├── go.mod
│   └── README.md
├── .gitignore
└── README.md
```

## Documentation

- [Server commands and HTTP API](server/README.md)
- [Architecture](docs/architecture.md)
- [Gameplay rules](docs/gameplay.md)

*Made with curiosity, questionable directions, and probably a few wrong turns.*
