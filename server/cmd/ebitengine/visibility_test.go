package main

import (
	"testing"

	"github.com/putriindaht/leftnt/server/internal/domain"
)

func TestEntitiesUseTheSameVisibilityAsTerrain(t *testing.T) {
	maze := domain.Maze{Cells: [][]bool{
		{false, false, false, false, false},
		{false, true, false, true, false},
		{false, true, true, true, false},
		{false, true, true, true, false},
		{false, false, false, false, false},
	}}
	tests := []struct {
		name    string
		target  domain.Position
		visible bool
	}{
		{"nearby open path", domain.Position{X: 1, Y: 2}, true},
		{"first blocking wall", domain.Position{X: 2, Y: 1}, true},
		{"goal behind a wall", domain.Position{X: 3, Y: 1}, false},
		{"goal past a blocked corner", domain.Position{X: 2, Y: 2}, false},
		{"goal outside viewing radius", domain.Position{X: 3, Y: 3}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := withinSight(maze, test.target, 1, 1, 2); got != test.visible {
				t.Errorf("expected visibility %t, got %t", test.visible, got)
			}
		})
	}
}

func TestSpeechBubblesStayInsideTheMaze(t *testing.T) {
	for _, message := range wallMessages {
		for _, position := range []domain.Position{{X: 1, Y: 1}, {X: 13, Y: 1}, {X: 1, Y: 9}, {X: 13, Y: 9}, {X: 7, Y: 5}} {
			layout := layoutWallBubble(message, float64(position.X), float64(position.Y))
			box := layout.bounds
			if box.Min.X < 4 || box.Max.X > mazeWidth*tileSize-4 || box.Min.Y < 4 || box.Max.Y > mazeHeight*tileSize-4 {
				t.Fatalf("bubble for %q at %v escapes the board: %v", message, position, box)
			}
			if layout.tailX <= box.Min.X || layout.tailX >= box.Max.X {
				t.Fatal("bubble tail should stay attached to the box")
			}
			if box.Min.Y < (position.Y+1)*tileSize && box.Max.Y > position.Y*tileSize {
				t.Fatal("bubble should not cover the player's row")
			}
		}
	}
}

func TestLightExpandsViewWithoutRevealingThroughWalls(t *testing.T) {
	maze := domain.Maze{Cells: [][]bool{
		{false, false, false, false, false, false},
		{false, true, true, true, true, false},
		{false, false, false, false, false, false},
	}}
	far := domain.Position{X: 4, Y: 1}
	if withinSight(maze, far, 1, 1, 1) || !withinSight(maze, far, 1, 1, 3) {
		t.Fatal("radius three should reveal a distant corridor hidden at radius one")
	}
	maze.Cells[1][2] = false
	if withinSight(maze, far, 1, 1, 3) {
		t.Fatal("a lantern should never reveal the friend or pickups through walls")
	}
}
