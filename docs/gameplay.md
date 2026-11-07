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
