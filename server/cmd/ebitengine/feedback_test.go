package main

import (
	"strings"
	"testing"
	"time"

	"github.com/putriindaht/leftnt/server/internal/domain"
)

func TestWallMessagesDoNotRepeatConsecutively(t *testing.T) {
	previous := ""
	for i := 0; i < 100; i++ {
		message := nextWallMessage(previous)
		if message == previous {
			t.Fatal("consecutive wall messages should differ")
		}
		found := false
		for _, candidate := range wallMessages {
			if message == candidate {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("unexpected wall message %q", message)
		}
		previous = message
	}
}

func TestWallFeedbackCooldownAndExpiry(t *testing.T) {
	feedback := wallFeedback{}
	feedback.bump()
	first := feedback.message
	feedback.bump()
	if feedback.message != first || feedback.remaining != wallMessageDuration {
		t.Fatal("cooldown should preserve the visible message")
	}
	for i := 0; i < wallMessageCooldown; i++ {
		feedback.advance()
	}
	feedback.bump()
	if feedback.message == first {
		t.Fatal("a new allowed bump should select a different message")
	}
	for i := 0; i < wallMessageDuration; i++ {
		feedback.advance()
	}
	if feedback.remaining != 0 || feedback.cooldown != 0 {
		t.Fatal("message and cooldown should expire")
	}
	feedback.advance()
	if feedback.remaining < 0 || feedback.cooldown < 0 {
		t.Fatal("timers must not become negative")
	}
}

func TestHoldingAgainstWallDoesNotRefreshFeedback(t *testing.T) {
	state := &domain.SoloGame{Game: &domain.Game{Maze: domain.Maze{Cells: [][]bool{{false, false, false}, {false, true, false}, {false, false, false}}}, Player: domain.Player{Position: domain.Position{X: 1, Y: 1}}, Status: domain.GameStatusPlaying}}
	game := &Game{state: state, walk: newWalkAnimation(state.Player.Position)}
	if err := game.startMove("up"); err != nil {
		t.Fatal(err)
	}
	first := game.feedback.message
	if first == "" {
		t.Fatal("the first blocked move should show feedback")
	}
	for i := 0; i < wallMessageCooldown; i++ {
		game.feedback.advance()
	}
	remaining := game.feedback.remaining
	for i := 0; i < 100; i++ {
		if err := game.startMove("up"); err != nil {
			t.Fatal(err)
		}
	}
	if game.feedback.message != first || game.feedback.remaining != remaining {
		t.Fatal("holding against the same wall must not refresh or replace the message")
	}
	if game.state.Moves != 0 || game.walk.moving() {
		t.Fatal("blocked attempts must not count as moves or start walking")
	}
	game.blockedDirection = ""
	if err := game.startMove("up"); err != nil {
		t.Fatal(err)
	}
	if game.feedback.message == first {
		t.Fatal("a fresh attempt should allow new feedback after cooldown")
	}
}

func TestMessagesFitTwoBubbleLines(t *testing.T) {
	messages := append([]string{}, wallMessages...)
	messages = append(messages, "Arrows/WASD: move. T: change walls.", "Your sense of direction survived. Try a fresh maze?")
	for _, message := range messages {
		wrapped := wrapMessage(message, bubbleColumns)
		lines := strings.Split(wrapped, "\n")
		if len(lines) > 2 {
			t.Errorf("message exceeds two bubble lines: %q", message)
		}
		for _, line := range lines {
			if len(line) > bubbleColumns {
				t.Errorf("bubble line is too wide: %q", line)
			}
		}
		if strings.Join(strings.Fields(wrapped), " ") != message {
			t.Errorf("wrapping changed message content: %q", message)
		}
	}
}

func TestResetStateClearsThePreviousRun(t *testing.T) {
	oldState := &domain.SoloGame{Game: &domain.Game{Status: domain.GameStatusWon, Moves: 42, Player: domain.Player{Position: domain.Position{X: 5, Y: 5}}}, LightRemaining: 5 * time.Second, LanternsCollected: 2}
	game := &Game{state: oldState, useHedge: false, queuedDirection: "right", blockedDirection: "up", winAnnounced: true, feedback: wallFeedback{message: "Old message", remaining: 90, cooldown: 10}, walk: newWalkAnimation(oldState.Player.Position)}
	game.walk.begin(domain.Position{X: 4, Y: 5}, oldState.Player.Position)
	newState, err := domain.NewSoloGame(15, 11, domain.DefaultSoloConfig())
	if err != nil {
		t.Fatal(err)
	}
	game.resetState(newState)
	if game.state != newState || game.state.Status != domain.GameStatusPlaying || game.state.Moves != 0 {
		t.Fatal("restart must install a fresh playing state with zero moves")
	}
	x, y := game.walk.position()
	if game.walk.moving() || x != 1 || y != 1 {
		t.Fatal("restart must place the visual player at the start without an old animation")
	}
	if game.queuedDirection != "" || game.blockedDirection != "" || game.winAnnounced || game.feedback.message != "" || game.feedback.remaining != 0 || game.feedback.cooldown != 0 {
		t.Fatal("restart must clear queued input, winning flags, and previous feedback")
	}
	if game.state.LightRemaining != 0 || game.state.LanternsCollected != 0 || game.state.VisionRadius() != 1 || len(game.state.Lanterns) != 2 {
		t.Fatal("restart must reset light and replace the lantern pickups")
	}
	if game.useHedge {
		t.Fatal("restart must preserve the selected wood wall style")
	}
}

func TestRestartButtonBoundsStayInsideHUD(t *testing.T) {
	bounds := restartBounds()
	if bounds.Min.X < 0 || bounds.Max.X > mazeWidth*tileSize || bounds.Min.Y < mazeHeight*tileSize || bounds.Max.Y > mazeHeight*tileSize+hudHeight {
		t.Fatal("restart button must fit inside the HUD, not the maze")
	}
}
