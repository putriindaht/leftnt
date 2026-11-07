package domain

import "testing"

func TestMazeIsWalkable(t *testing.T) {
	maze := Maze{
		Cells: [][]bool{
			{false, false, false},
			{false, true, false},
			{false, false, false},
		},
	}

	if !maze.IsWalkable(Position{X: 1, Y: 1}) {
		t.Errorf("a path cell should be walkable")
	}

	if maze.IsWalkable(Position{X: 0, Y: 1}) {
		t.Error("a wall cell should not be walkable")
	}

	if maze.IsWalkable(Position{X: -1, Y: 1}) {
		t.Error("a negative column should not be walkable")
	}

	if maze.IsWalkable(Position{X: 1, Y: 3}) {
		t.Error("a row outside the maze should not be walkable")
	}

}
