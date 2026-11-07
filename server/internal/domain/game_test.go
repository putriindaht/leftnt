package domain

import (
	"errors"
	"testing"
)

func TestNewGame(t *testing.T) {
	game, err := NewGame(15, 11)
	if err != nil {
		t.Fatal(err)
	}
	if game.Status != GameStatusPlaying || game.Moves != 0 || game.Player.Position != (Position{X: 1, Y: 1}) {
		t.Fatalf("unexpected initial game: %+v", game)
	}
	distance, reachable := game.Maze.distancesFrom(game.Player.Position)[game.Goal]
	if !reachable || distance == 0 {
		t.Fatal("goal must be reachable and different from the start")
	}
}

func TestNewGameRejectsInvalidDimensions(t *testing.T) {
	for _, size := range [][2]int{{0, 11}, {14, 11}, {53, 11}, {15, 52}, {3, 3}} {
		if _, err := NewGame(size[0], size[1]); !errors.Is(err, ErrInvalidDimensions) {
			t.Errorf("size %v: expected dimension error, got %v", size, err)
		}
	}
}

func TestGameMoveAndWin(t *testing.T) {
	game := Game{Maze: mazeFromLayout(t, "#####", "#...#", "#####"), Player: Player{Position: Position{X: 1, Y: 1}}, Goal: Position{X: 3, Y: 1}, Status: GameStatusPlaying}
	if moved, err := game.Move("up"); moved || err != nil || game.Moves != 0 {
		t.Fatal("blocked moves should leave the game unchanged")
	}
	if _, err := game.Move("teleport"); !errors.Is(err, ErrInvalidDirection) {
		t.Fatal("invalid directions should be rejected")
	}
	if moved, err := game.Move("right"); !moved || err != nil || game.Status != GameStatusPlaying {
		t.Fatal("first valid move should keep the game playing")
	}
	if moved, err := game.Move("right"); !moved || err != nil || game.Status != GameStatusWon || game.Moves != 2 {
		t.Fatal("reaching the goal should win in two moves")
	}
	if _, err := game.Move("left"); !errors.Is(err, ErrGameFinished) || game.Player.Position != game.Goal || game.Moves != 2 {
		t.Fatal("finished games must not change")
	}
}
