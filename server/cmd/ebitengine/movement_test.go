package main

import (
	"errors"
	"math"
	"testing"

	"github.com/putriindaht/leftnt/server/internal/domain"
)

func TestWalkAnimationInterpolatesEachDirection(t *testing.T) {
	start := domain.Position{X: 2, Y: 2}
	tests := []struct {
		name                 string
		target               domain.Position
		midpointX, midpointY float64
	}{
		{"right", domain.Position{X: 3, Y: 2}, 2.5, 2},
		{"left", domain.Position{X: 1, Y: 2}, 1.5, 2},
		{"up", domain.Position{X: 2, Y: 1}, 2, 1.5},
		{"down", domain.Position{X: 2, Y: 3}, 2, 2.5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			walk := newWalkAnimation(start)
			if walk.moving() || walk.frame() != 0 {
				t.Fatal("new animation should be idle")
			}
			walk.begin(start, test.target)
			x, y := walk.position()
			if x != 2 || y != 2 {
				t.Fatal("a step should start at the previous tile")
			}
			for i := 0; i < walkDuration/2; i++ {
				walk.advance()
			}
			x, y = walk.position()
			if math.Abs(x-test.midpointX) > 0.00001 || math.Abs(y-test.midpointY) > 0.00001 {
				t.Fatalf("expected midpoint (%v,%v), got (%v,%v)", test.midpointX, test.midpointY, x, y)
			}
			if !walk.moving() {
				t.Fatal("the halfway point should still be moving")
			}
			for i := 0; i < walkDuration; i++ {
				walk.advance()
			}
			if walk.moving() || walk.frame() != 0 {
				t.Fatal("completed animation should return to idle")
			}
			x, y = walk.position()
			if x != float64(test.target.X) || y != float64(test.target.Y) {
				t.Fatal("animation should stop exactly at its target without overshooting")
			}
		})
	}
}

func TestWalkAnimationAlternatesFeet(t *testing.T) {
	walk := newWalkAnimation(domain.Position{X: 1, Y: 1})
	walk.begin(domain.Position{X: 1, Y: 1}, domain.Position{X: 2, Y: 1})
	for i := 0; i < walkDuration/2; i++ {
		walk.advance()
	}
	if walk.frame() != 1 {
		t.Fatal("first step should use the left-foot frame")
	}
	for walk.moving() {
		walk.advance()
	}
	walk.begin(domain.Position{X: 2, Y: 1}, domain.Position{X: 3, Y: 1})
	for i := 0; i < walkDuration/2; i++ {
		walk.advance()
	}
	if walk.frame() != 2 {
		t.Fatal("second step should use the right-foot frame")
	}
}

func TestMovementCollisionAndDelayedRescue(t *testing.T) {
	state := &domain.SoloGame{Game: &domain.Game{
		Maze:   domain.Maze{Cells: [][]bool{{false, false, false, false}, {false, true, true, false}, {false, false, false, false}}},
		Player: domain.Player{Position: domain.Position{X: 1, Y: 1}},
		Goal:   domain.Position{X: 2, Y: 1}, Status: domain.GameStatusPlaying,
	}}
	game := &Game{state: state, walk: newWalkAnimation(state.Player.Position)}
	if err := game.startMove("up"); err != nil {
		t.Fatal(err)
	}
	if game.walk.moving() || state.Moves != 0 {
		t.Fatal("walking into a wall must not animate or count as a move")
	}
	if tileVisible(10, 10, 1, 1, 1, game.friendFound()) {
		t.Fatal("distant cells should be hidden before finding the friend")
	}
	if err := game.startMove("right"); err != nil {
		t.Fatal(err)
	}
	if state.Status != domain.GameStatusWon {
		t.Fatal("domain should register reaching the goal")
	}
	if game.friendFound() {
		t.Fatal("the reveal must wait for the visual step to finish")
	}
	for i := 0; i < walkDuration-1; i++ {
		game.walk.advance()
	}
	if game.friendFound() {
		t.Fatal("the reveal must not happen one tick early")
	}
	game.walk.advance()
	if !game.friendFound() {
		t.Fatal("the friend should be found on arrival")
	}
	if !tileVisible(10, 10, 2, 1, 1, game.friendFound()) {
		t.Fatal("the entire map should be revealed after finding the friend")
	}
	if err := game.startMove("left"); !errors.Is(err, domain.ErrGameFinished) {
		t.Fatal("the domain must reject further movement after winning")
	}
}

func TestVisibilityFollowsTheAnimatedPosition(t *testing.T) {
	if tileVisible(3, 1, 1, 1, 1, false) {
		t.Fatal("cell should start outside the radius")
	}
	if !tileVisible(3, 1, 2, 1, 1, false) {
		t.Fatal("cell should enter view as the player moves")
	}
	if !tileVisible(1, 1, 1, 1, 1, false) {
		t.Fatal("the player's cell must be visible")
	}
}
