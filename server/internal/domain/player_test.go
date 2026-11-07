package domain

import "testing"

func TestPlayerMove(t *testing.T) {
	maze := Maze{
		Cells: [][]bool{
			{false, false, false, false, false},
			{false, true, true, false, false},
			{false, false, false, false, false},
		},
	}

	player := Player{
		Position: Position{X: 1, Y: 1},
	}

	moved := player.Move("right", maze)

	if !moved {
		t.Fatal("moving onto a path should succeed")
	}

	expected := Position{X: 2, Y: 1}

	if player.Position != expected {
		t.Fatalf("expected position %v, got %v", expected, player.Position)
	}

	moved = player.Move("right", maze)

	if moved {
		t.Error("moving into a wall should fail")
	}

	if player.Position != expected {
		t.Errorf("expected position to remain %v, got %v", expected, player.Position)
	}
}
