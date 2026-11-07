package domain

import (
	"errors"
	"testing"
	"time"
)

func soloFixture(t *testing.T) *SoloGame {
	t.Helper()
	return &SoloGame{
		Game: &Game{
			Maze:   mazeFromLayout(t, "#########", "#.......#", "#########"),
			Player: Player{Position: Position{X: 1, Y: 1}},
			Goal:   Position{X: 7, Y: 1}, Status: GameStatusPlaying,
		},
		Lanterns: []Position{{X: 2, Y: 1}, {X: 4, Y: 1}},
		config:   DefaultSoloConfig(),
	}
}

func TestSoloLanternPickupRefillAndExpiry(t *testing.T) {
	game := soloFixture(t)
	if game.VisionRadius() != 1 || game.LightRemaining != 0 {
		t.Fatal("solo should start with radius one and no active light")
	}
	if moved, err := game.Move("up"); moved || err != nil || game.LightRemaining != 0 || len(game.Lanterns) != 2 {
		t.Fatal("blocked movement must not collect a lantern")
	}
	if moved, err := game.Move("right"); !moved || err != nil {
		t.Fatal("player should reach the first lantern")
	}
	if game.Status != GameStatusPlaying || game.VisionRadius() != 3 || game.LightRemaining != 20*time.Second || len(game.Lanterns) != 1 || game.LanternsCollected != 1 {
		t.Fatal("a pickup should grant twenty seconds of light without winning")
	}
	game.AdvanceLight(7 * time.Second)
	if game.LightRemaining != 13*time.Second {
		t.Fatal("light should decrease even when the player is not moving")
	}
	for _, direction := range []string{"left", "right"} {
		if _, err := game.Move(direction); err != nil {
			t.Fatal(err)
		}
	}
	if game.LightRemaining != 13*time.Second || game.LanternsCollected != 1 {
		t.Fatal("revisiting a collected tile must not refill the lantern")
	}
	for i := 0; i < 2; i++ {
		if _, err := game.Move("right"); err != nil {
			t.Fatal(err)
		}
	}
	if game.LightRemaining != 20*time.Second || game.LanternsCollected != 2 || len(game.Lanterns) != 0 {
		t.Fatal("a second pickup should refill to twenty seconds, not stack time")
	}
	game.AdvanceLight(-time.Second)
	if game.LightRemaining != 20*time.Second {
		t.Fatal("negative elapsed time must be ignored")
	}
	game.AdvanceLight(20 * time.Second)
	if game.LightRemaining != 0 || game.VisionRadius() != 1 {
		t.Fatal("the light should expire exactly at its duration and restore radius one")
	}
	game.AdvanceLight(time.Second)
	if game.LightRemaining != 0 {
		t.Fatal("the timer must never go below zero")
	}
	for i := 0; i < 3; i++ {
		if _, err := game.Move("right"); err != nil {
			t.Fatal(err)
		}
	}
	if game.Status != GameStatusWon {
		t.Fatal("reaching the friend should win, even without light")
	}
	if _, err := game.Move("left"); !errors.Is(err, ErrGameFinished) {
		t.Fatal("winning must stop movement")
	}
}

func TestSoloTimerStopsWhenFriendIsFound(t *testing.T) {
	game := soloFixture(t)
	for game.Status == GameStatusPlaying {
		if _, err := game.Move("right"); err != nil {
			t.Fatal(err)
		}
	}
	remaining := game.LightRemaining
	game.AdvanceLight(time.Hour)
	if game.LightRemaining != remaining {
		t.Fatal("the timer should freeze after winning")
	}
}

func TestSoloLanternsAreReachableAndSpreadAlongTheRoute(t *testing.T) {
	maze := mazeFromLayout(t, "###############################", "#.............................#", "###############################")
	start, goal := Position{X: 1, Y: 1}, Position{X: 29, Y: 1}
	for attempt := 0; attempt < 30; attempt++ {
		lanterns := maze.randomLanterns(start, goal, 2)
		if len(lanterns) != 2 || lanterns[0].X >= 15 || lanterns[1].X < 15 || lanterns[1].X-lanterns[0].X < 7 {
			t.Fatalf("pickups must occupy separated route sections: %v", lanterns)
		}
	}
	for attempt := 0; attempt < 20; attempt++ {
		game, err := NewSoloGame(15, 11, DefaultSoloConfig())
		if err != nil {
			t.Fatal(err)
		}
		if len(game.Lanterns) != 2 {
			t.Fatal("the default maze should have two lanterns")
		}
		reachable := game.Maze.distancesFrom(game.Player.Position)
		seen := map[Position]bool{}
		for _, position := range game.Lanterns {
			_, ok := reachable[position]
			if !ok || seen[position] || position == game.Goal || position == game.Player.Position {
				t.Fatalf("invalid lantern position: %v", position)
			}
			seen[position] = true
		}
	}
}

func TestSoloSettingsAndSmallMazes(t *testing.T) {
	config := DefaultSoloConfig()
	config.LanternCount = 4
	config.LanternDuration = 5 * time.Second
	game, err := NewSoloGame(15, 11, config)
	if err != nil {
		t.Fatal(err)
	}
	if len(game.Lanterns) != 4 {
		t.Fatal("the configured lantern count should be used")
	}
	fixture := soloFixture(t)
	fixture.config = config
	if _, err := fixture.Move("right"); err != nil {
		t.Fatal(err)
	}
	if fixture.LightRemaining != 5*time.Second {
		t.Fatal("the configured light duration should be used")
	}
	config.LanternCount = 0
	game, err = NewSoloGame(15, 11, config)
	if err != nil || len(game.Lanterns) != 0 {
		t.Fatal("zero lanterns should be a supported configuration")
	}
	config.LanternCount = 10
	game, err = NewSoloGame(3, 5, config)
	if err != nil || len(game.Lanterns) != 1 {
		t.Fatal("tiny mazes should use available cells without duplicates")
	}
	for _, change := range []func(*SoloConfig){
		func(c *SoloConfig) { c.LanternCount = -1 },
		func(c *SoloConfig) { c.LanternDuration = 0 },
		func(c *SoloConfig) { c.BaseVisionRadius = 0 },
		func(c *SoloConfig) { c.LanternVisionRadius = 1 },
	} {
		invalid := DefaultSoloConfig()
		change(&invalid)
		if _, err := NewSoloGame(15, 11, invalid); err == nil {
			t.Fatal("invalid solo settings must be rejected")
		}
	}
	if _, err := NewSoloGame(4, 5, DefaultSoloConfig()); !errors.Is(err, ErrInvalidDimensions) {
		t.Fatal("solo should preserve shared maze dimension validation")
	}
}
