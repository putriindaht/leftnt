# Game HTTP API

The Go server listens on `http://localhost:8080`. The Next.js client forwards
`/api/*` requests to this server. Direct clients such as curl use the paths below.
JSON responses use `Cache-Control: no-store` for game state.

## Create a game

```sh
curl -i http://localhost:8080/games \
  -H 'Content-Type: application/json' \
  -d '{"width":15,"height":11}'
```

Returns `201 Created` with a `Location: /games/{id}` header and game state.
An empty object uses width 15 and height 11. Dimensions must be odd, between 3
and 51 inclusive, and at least one dimension must exceed 3 so a distinct goal
exists. Invalid dimensions return `400`.

## Read a game

```sh
curl http://localhost:8080/games/GAME_ID
```

Returns `200` and game state, or `404` if the ID does not exist.

## Move a player

```sh
curl http://localhost:8080/games/GAME_ID/moves \
  -H 'Content-Type: application/json' \
  -d '{"direction":"right"}'
```

Directions are `up`, `down`, `left`, and `right`. Returns `200` with updated
state and a `moved` boolean. A blocked move returns `moved: false`, leaving the
position and successful move count unchanged. A move onto the goal sets status
to `won`. Further movement requests return `409`.

## Response shape

Every successful game response contains:

- `id`: server-generated identifier.
- `maze`: `width`, `height`, and `cells` (boolean rows).
- `player`: `{ "x": number, "y": number }`.
- `goal`: `{ "x": number, "y": number }`.
- `status`: `playing` or `won`.
- `moves`: count of successful moves, initially zero.
- `moved`: boolean present only in a move response.

The full maze is returned. Fog of war and multiplayer roles are not implemented.

## Errors and lifecycle

Errors are JSON objects with an `error` string. Invalid JSON, unknown fields,
unsupported directions, and invalid dimensions return `400`; missing games
return `404`; finished games return `409`. Request bodies are limited to 4 KiB.

This is a local single-player prototype. Games live in one process's memory
and disappear when the API restarts. Updates are serialized so simultaneous
requests cannot overwrite each other's moves. Anyone with a game ID can read
or move that game; there are no accounts or player authorization yet. There is
no automatic game expiration in this prototype.
