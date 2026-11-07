# Architecture

The server separates startup, HTTP delivery, application operations, and game
rules. Packages are added when they have a concrete responsibility.

| Path | Responsibility |
| --- | --- |
| `server/cmd/api` | Construct the health service and router; start the HTTP server. |
| `server/cmd/play` | Read terminal input, render the maze, and detect a completed game. |
| `server/internal/delivery/http` | Match HTTP routes and translate application results into JSON. |
| `server/internal/usecase` | Provide application operations; health and game services. |
| `server/internal/storage/memory` | Store game snapshots and apply atomic updates under a mutex. |
| `client` | Render game state with Next.js and send typed HTTP requests. |
| `server/internal/domain` | Define positions, maze generation, walkability, movement, and goal selection. |

## Dependencies

HTTP delivery depends on the use case package. The terminal entry point depends
directly on the domain package. Game use cases depend on a repository interface;
the in-memory storage adapter implements it and depends inward on domain/usecase. Both entry points construct and connect the
objects they need. The domain package does not depend on HTTP or terminal I/O;
the use case package does not depend on HTTP delivery. Go's `internal` directory
limits package access to code within the parent module tree.

There are no persistence adapters or database dependencies at this stage.

## HTTP request flow

`cmd/api` connects `HealthService` to the router. A `GET /health` request reaches
`HealthHandler.Check`, which calls the service through `HealthChecker` and
writes JSON. Response headers and JSON field names belong to the HTTP layer.

## Terminal flow

`cmd/play` generates a maze, creates a player, and chooses a goal before the
input loop. The terminal input adapter translates WASD and arrow-key escape sequences into
direction names. Each movement command calls `Game.Move` and redraws the board.
If the player reaches the goal, the program prints a win message and exits.
The player and goal are stored as separate positions, not special grid values.

Each executable stores state in memory. The terminal game and HTTP server are
independent processes: they do not share a game session or communicate.
Restarting the terminal game creates a fresh maze.

## Browser game flow

Next.js forwards `/api/games` to the Go API. Game handlers validate JSON and call
the game service. The service creates domain games or requests an atomic update
from its repository. Domain `Game.Move` owns collision, move count, and winning.
Responses contain explicit JSON views, not raw exported domain field names.

Browser and terminal games use the same `NewGame` and `Game.Move` rules. API games
have server-generated IDs and remain in memory until process restart. Repository
reads return isolated snapshots; updates use a mutex to avoid lost moves.
