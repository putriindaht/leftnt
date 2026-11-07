# Gameplay reference

## Coordinates and grid

Positions use `(X, Y)`: X is the column and Y is the row. Grid access uses
`Cells[Y][X]`. Indices start at zero. X increases right; Y increases down.
A cell is `true` for a path or `false` for a wall. The player and goal are stored
separately and drawn over the grid.

## Maze generation

`NewMaze` accepts odd widths and heights of at least 3. It initializes walls
and opens `(1, 1)`. `GenerateMaze` uses randomized depth-first search with a
stack. It chooses an unvisited interior destination two cells away, opens the
destination and intervening wall, and backtracks when no candidate remains.
Outer borders stay closed; every generated path cell is connected to the start.
Generation advances two cells at a time; the player moves one cell at a time.

The API and terminal both use `Game.Move` to apply these rules. Only successful
moves increment the move count. A game is `playing` until reaching the goal,
then becomes `won` and rejects further moves.

## Movement

| Command | Direction | Coordinate change |
| --- | --- | --- |
| `W` or ↑ | Up | Y − 1 |
| `S` or ↓ | Down | Y + 1 |
| `A` or ← | Left | X − 1 |
| `D` or → | Right | X + 1 |

Arrow keys and WASD move immediately without Enter in an interactive terminal.
Both letter cases are accepted. Press Q or Ctrl+C to quit.
The player moves only when the destination is walkable.
Walls, out-of-bounds positions, and unrecognized directions leave it unchanged.

## Goal selection

`RandomGoal` uses breadth-first search to find shortest walking distances from
the start. Candidates are reachable path cells other than the start satisfying
`distance * 10 >= farthestDistance * 7`: at least 70% of the farthest distance.
Whole-step distances effectively round fractional thresholds up.

This is a relative threshold, not an absolute step minimum. If the only other
reachable cell is one step away, it is valid. An invalid start or the absence
of another reachable path cell returns an error.

The goal is an interior destination, not an opening in the border. Reaching it
ends the game. Random outcomes may repeat across runs.

## Display

| Symbol | Meaning |
| --- | --- |
| `P` | Player |
| `G` | Goal |
| `.` | Path |
| `#` | Wall |

When the player reaches the goal, `P` takes display priority over `G`.

## Ebitengine presentation

The Ebitengine solo mode uses `SoloGame` to extend the shared collision and goal
rules. The goal represents the explorer's lost friend, a boy in a red shirt.
Visual movement
interpolates between accepted grid positions over 12 ticks; the domain still
stores integer positions. Arrow keys/WASD can be held for continuous movement.
The player uses idle and alternating foot poses, currently front-facing.

Before winning, only cells within a one-cell radius of the animated player
are drawn. Collecting a lantern expands this to three cells for 20 seconds of
game time, including time spent standing still. Another pickup refills that
duration rather than adding to it. After expiry, the radius returns to one.
Two lanterns are randomly spread across different sections of the route from
the start to the friend. Each pickup is consumed once. `DefaultSoloConfig`
controls the count, duration, and view radii; tiny mazes use available cells.

Visible terrain is dim and cool without light, then brighter and warmer while
a lantern is active. Characters, pickups, and UI retain their normal colors.
Winning restores the map's full brightness.

Walls also block sight, including corners. The friend and lanterns share
terrain visibility, so neither is shown through fog or walls. Reaching the
friend completes the final animated step, displays the characters together,
reveals the map, and freezes the light timer. T swaps hedge and wood textures
without changing maze geometry. This presentation does not connect to the HTTP API.

The compact Ebitengine HUD below the maze separates the objective or win message,
move/pickup counts, light timer, view radius, and controls onto their own lines. Restart
with R or the on-screen button at any time; this generates a new game, resets
moves, pickups, light duration, animation, queued input, visibility state, and feedback, and preserves
the selected hedge/wood style. Blocked moves produce a random playful speech
bubble near the player without changing position or move count. Bubbles stay
inside the maze and appear above or below the player to keep the character clear.
The same line is not selected twice
in a row, and sustained pressure against a wall does not spam new messages.
