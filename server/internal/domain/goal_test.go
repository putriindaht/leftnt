package domain

import "testing"

func mazeFromLayout(t *testing.T, rows ...string) Maze {
	t.Helper()
	maze := Maze{Cells: make([][]bool, len(rows))}
	for y, row := range rows {
		maze.Cells[y] = make([]bool, len(row))
		for x, cell := range row {
			switch cell {
			case '.':
				maze.Cells[y][x] = true
			case '#':
			default:
				t.Fatalf("unexpected maze symbol %q at (%d, %d)", cell, x, y)
			}
		}
	}
	return maze
}

func windingMaze(t *testing.T) Maze {
	t.Helper()
	return mazeFromLayout(t,
		"#########",
		"#.#...#.#",
		"#.###.###",
		"#.....###",
		"#########",
	)
}

func TestDistancesFromFollowsCorridors(t *testing.T) {
	maze := windingMaze(t)
	distances := maze.distancesFrom(Position{X: 1, Y: 1})

	// The ordered path takes a detour around the wall.
	path := []Position{
		{X: 1, Y: 1}, {X: 1, Y: 2}, {X: 1, Y: 3},
		{X: 2, Y: 3}, {X: 3, Y: 3}, {X: 4, Y: 3},
		{X: 5, Y: 3}, {X: 5, Y: 2}, {X: 5, Y: 1},
		{X: 4, Y: 1}, {X: 3, Y: 1},
	}

	if len(distances) != len(path) {
		t.Fatalf("expected %d reachable cells, got %d", len(path), len(distances))
	}
	for expected, position := range path {
		actual, found := distances[position]
		if !found || actual != expected {
			t.Errorf("distance to %v: expected %d, got %d (found: %t)", position, expected, actual, found)
		}
	}
	if _, found := distances[Position{X: 7, Y: 1}]; found {
		t.Error("an isolated path cell should not be reachable")
	}
}

func TestDistancesFromFindsShortestRoute(t *testing.T) {
	maze := mazeFromLayout(t,
		"#####",
		"#...#",
		"#.#.#",
		"#...#",
		"#####",
	)
	distances := maze.distancesFrom(Position{X: 1, Y: 1})
	if len(distances) != 8 {
		t.Fatalf("expected 8 reachable cells in the loop, got %d", len(distances))
	}
	if distance := distances[Position{X: 3, Y: 1}]; distance != 2 {
		t.Errorf("expected the shortest route to take 2 steps, got %d", distance)
	}
	if distance := distances[Position{X: 3, Y: 3}]; distance != 4 {
		t.Errorf("expected the opposite corner to take 4 steps, got %d", distance)
	}
}

func TestRandomGoalSelectsDistantReachableCell(t *testing.T) {
	maze := windingMaze(t)
	start := Position{X: 1, Y: 1}

	// The farthest cell is 10 steps away, so eligible distances are 7–10.
	allowed := map[Position]bool{
		{X: 5, Y: 2}: true,
		{X: 5, Y: 1}: true,
		{X: 4, Y: 1}: true,
		{X: 3, Y: 1}: true,
	}

	// Check the selection rule, without requiring random results to differ.
	for attempt := 0; attempt < 50; attempt++ {
		goal, err := maze.RandomGoal(start)
		if err != nil {
			t.Fatalf("failed to select goal: %v", err)
		}
		if !allowed[goal] {
			t.Fatalf("selected ineligible goal %v; expected a reachable cell 7–10 steps away", goal)
		}
	}
}

func TestRandomGoalDoesNotRoundMinimumDistanceDown(t *testing.T) {
	maze := mazeFromLayout(t, "######", "#....#", "######")
	goal, err := maze.RandomGoal(Position{X: 1, Y: 1})
	if err != nil {
		t.Fatalf("failed to select goal: %v", err)
	}
	// 70% of 3 is 2.1, so a goal 2 steps away is too close.
	expected := Position{X: 4, Y: 1}
	if goal != expected {
		t.Errorf("expected goal %v at distance 3, got %v", expected, goal)
	}
}

func TestRandomGoalAllowsSingleStepWhenItIsTheOnlyOption(t *testing.T) {
	maze := mazeFromLayout(t, "####", "#..#", "####")
	goal, err := maze.RandomGoal(Position{X: 1, Y: 1})
	if err != nil {
		t.Fatalf("failed to select goal: %v", err)
	}
	if goal != (Position{X: 2, Y: 1}) {
		t.Errorf("expected the only other path cell, got %v", goal)
	}
}

func TestRandomGoalReturnsError(t *testing.T) {
	tests := []struct {
		name  string
		maze  Maze
		start Position
	}{
		{name: "empty maze", maze: Maze{}, start: Position{}},
		{name: "start on wall", maze: mazeFromLayout(t, "###", "#.#", "###"), start: Position{}},
		{name: "start outside maze", maze: mazeFromLayout(t, "###", "#.#", "###"), start: Position{X: -1, Y: 1}},
		{name: "no other reachable cell", maze: mazeFromLayout(t, "#####", "#.#.#", "#####"), start: Position{X: 1, Y: 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.maze.RandomGoal(test.start)
			if err == nil {
				t.Error("expected goal selection to fail")
			}
		})
	}
}
