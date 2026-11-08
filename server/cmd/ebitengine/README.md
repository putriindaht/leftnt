# Ebitengine game

A local, single-player rescue game using `domain.SoloGame`, which extends the
shared maze and movement rules. The player is a blue-jacketed explorer with
brown pigtails and pink hair ties. Find her lost friend: a short-haired boy
in a red shirt. Lantern pickups temporarily improve visibility. Sprites and
terrain are generated from code.

## Run

From `server`:

```sh
go run ./cmd/ebitengine
```

If the default macOS Metal backend produces rendering errors, compare with:

```sh
EBITENGINE_GRAPHICS_LIBRARY=opengl go run ./cmd/ebitengine
```

## Controls and behavior

- Arrow keys or WASD move one cell per step. Holding a key continues walking.
- T switches between hedge and wood walls. Hedge is the default.
- R or the on-screen Restart button creates a fresh maze without closing the game.
  The button accepts mouse clicks and touch. The selected wall style is preserved.
- Movement is checked by the shared domain before visual animation starts.
- A valid step takes 12 ticks (about 0.2 seconds at the default 60 ticks/second).
- Idle and two alternating foot poses are cached at startup. The poses currently
  face forward; directional facing sprites are not implemented.
- Start with a one-cell view radius. Walking onto a lantern automatically
  collects it and increases the radius to three for 20 seconds of game time.
  The timer runs while standing still and walking; it pauses when the engine
  is suspended, such as an unfocused window under the default settings.
- Another pickup refills the timer to 20 seconds. Time does not stack. When
  light expires, the radius returns to one. Each lantern can be collected once.
- Two lanterns are randomly placed in separate sections of the reachable route
  to the friend. Neither the starting cell nor the friend's cell contains one.
- The view follows the animated player position. Walls block sight, including
  corners. Terrain, pickups, and the friend share these visibility rules.
- Without light, visible paths and walls are dim with a cool tint. A lit
  lantern restores brighter, warmer terrain colors. Characters, pickups, and
  UI keep their normal colors. Rescuing the friend restores full map brightness.
- Reaching the friend completes the last visual step, reveals the entire maze,
  and stops movement and the light timer. Both characters stand together.
  The HUD announces the win on desktop and web; the desktop title also changes.
- A compact HUD separates the objective, move/pickup counts, light countdown,
  current view radius, and controls. Blocked movement displays a speech bubble near
  the player with one of 14 playful wall messages, excluding the previous one.
  Messages last about 3 seconds, with a 0.75-second cooldown. Holding against the
  same wall does not continuously replace the message.

Solo mode runs locally. Multiplayer uses server-owned Explorer/Navigator rooms
over WebSocket with filtered snapshots and the same pixel-art assets.
See [multiplayer](../../../docs/multiplayer.md) for browser and desktop setup.
The tracked `scripts/build-engine.sh` and Dockerfile build the browser launcher.
The separate original Next.js game continues using the single-player HTTP API.

## Solo settings

Edit `DefaultSoloConfig` in `internal/domain/solo.go` to adjust `LanternCount`,
`LanternDuration`, `BaseVisionRadius`, and `LanternVisionRadius`. Startup and
Restart both use these settings. Small mazes place as many lanterns as the
route's available cells allow. A count of zero is supported.

## Code

- `main.go`: keyboard input, rendering, and startup.
- `multiplayer.go`: remote snapshots, role-specific controls, and room HUD.
- `movement.go`: visual interpolation, animation frame selection, and visibility.
- `sprites.go`: player poses, the red-shirted friend, and lantern pixels.
- `tiles.go`: dirt, wood, and hedge textures.
- `hud.go`: win feedback, move count, text rendering, and the Restart button.
- `fonts.go`: a native 5×7 bitmap alphabet with solid pixels and no antialiasing.
- `display.go`: nearest-neighbor final rendering at whole window-pixel scales,
  accounting for high-DPI displays. Unused space is letterboxed. Windows smaller
  than the logical game size crop at 1× instead of shrinking the text.
- `bubble.go`: wall-message bubbles, text layout, and placement within the maze.
- `visibility.go`: shared sight checks for terrain, pickups, and the friend.
- `feedback.go`: random wall messages, expiry/cooldown, and text wrapping.
- `restart.go`: fresh game state while preserving cached images and wall style.
- `feedback_test.go`: message selection, cooldown, held-wall behavior, text fit,
  and state reset.
- `movement_test.go`: exact endpoints, midpoint positions, alternating feet,
  collision behavior, delayed rescue presentation, and visibility after winning.
- `visibility_test.go`: radius and wall occlusion, plus bubble placement at edges.

`internal/domain/solo.go` owns pickup placement, collection, light duration,
and view radii. It receives elapsed time from the engine and does not read a
clock itself. `solo_test.go` checks settings, placement, refill, expiry, and wins.

From `server`, run `go test ./...` for domain and renderer checks.
