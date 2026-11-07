package domain

import (
	"fmt"
	"testing"
)

func TestMazeConstructorsRejectInvalidDimensions(t *testing.T) {
	constructors := []struct {
		name string
		new  func(int, int) (Maze, error)
	}{
		{name: "NewMaze", new: NewMaze},
		{name: "GenerateMaze", new: GenerateMaze},
	}
	dimensions := [][2]int{{-1, 5}, {5, -1}, {0, 5}, {5, 0}, {1, 5}, {5, 1}, {4, 5}, {5, 4}}
	for _, constructor := range constructors {
		for _, size := range dimensions {
			t.Run(fmt.Sprintf("%s/%dx%d", constructor.name, size[0], size[1]), func(t *testing.T) {
				if _, err := constructor.new(size[0], size[1]); err == nil {
					t.Error("expected invalid dimensions to be rejected")
				}
			})
		}
	}
}

func TestNewMazeOpensOnlyStartingCell(t *testing.T) {
	maze, err := NewMaze(7, 5)
	if err != nil {
		t.Fatalf("failed to create maze: %v", err)
	}
	if len(maze.Cells) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(maze.Cells))
	}
	for y, row := range maze.Cells {
		if len(row) != 7 {
			t.Fatalf("row %d: expected 7 columns, got %d", y, len(row))
		}
		for x, walkable := range row {
			expected := x == 1 && y == 1
			if walkable != expected {
				t.Errorf("cell (%d, %d): expected walkable=%t, got %t", x, y, expected, walkable)
			}
		}
	}
}

func TestGenerateMazeProducesConnectedPathsAndClosedBorders(t *testing.T) {
	for _, size := range [][2]int{{3, 3}, {3, 7}, {7, 3}, {7, 5}, {15, 11}} {
		width, height := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			maze, err := GenerateMaze(width, height)
			if err != nil {
				t.Fatalf("failed to generate maze: %v", err)
			}
			if len(maze.Cells) != height {
				t.Fatalf("expected %d rows, got %d", height, len(maze.Cells))
			}
			reachable := maze.distancesFrom(Position{X: 1, Y: 1})
			for y, row := range maze.Cells {
				if len(row) != width {
					t.Fatalf("row %d: expected %d columns, got %d", y, width, len(row))
				}
				for x, walkable := range row {
					if (x == 0 || y == 0 || x == width-1 || y == height-1) && walkable {
						t.Errorf("border cell (%d, %d) should remain a wall", x, y)
					}
					if x%2 == 1 && y%2 == 1 && !walkable {
						t.Errorf("maze room (%d, %d) was not opened", x, y)
					}
					if walkable {
						if _, found := reachable[Position{X: x, Y: y}]; !found {
							t.Errorf("path cell (%d, %d) is unreachable from the start", x, y)
						}
					}
				}
			}
		})
	}
}
