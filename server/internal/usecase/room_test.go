package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/protocol"
	"github.com/putriindaht/leftnt/server/internal/storage/memory"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

func TestRoomRolesPrivacyLightReconnectAndRestart(t *testing.T) {
	repository := memory.NewRoomRepository()
	service := usecase.NewRoomService(repository)
	explorer, err := service.Create()
	if err != nil {
		t.Fatal(err)
	}
	err = repository.Update(explorer.Code, func(record *usecase.RoomRecord) error {
		cells := make([][]bool, 11)
		for y := range cells {
			cells[y] = make([]bool, 15)
		}
		for x := 1; x < 14; x++ {
			cells[1][x] = true
		}
		record.Room.Game.Game = &domain.Game{Maze: domain.Maze{Cells: cells}, Player: domain.Player{Position: domain.Position{X: 1, Y: 1}}, Goal: domain.Position{X: 7, Y: 1}, Status: domain.GameStatusPlaying}
		record.Room.Game.Lanterns = []domain.Position{{X: 2, Y: 1}, {X: 4, Y: 1}}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.Subscribe(explorer.Code, "invalid"); !errors.Is(err, usecase.ErrInvalidSession) {
		t.Fatal("invalid tokens must not subscribe")
	}
	explorerViews, leaveExplorer, err := service.Subscribe(explorer.Code, explorer.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveExplorer()
	waiting := <-explorerViews
	if waiting.Phase != "waiting" || waiting.Friend != nil || len(waiting.Lanterns) != 1 || waiting.Cells[1][4] != nil {
		t.Fatal("explorer must only receive nearby terrain and entities")
	}
	if _, _, err := service.Subscribe(explorer.Code, explorer.Token); !errors.Is(err, usecase.ErrAlreadyConnected) {
		t.Fatal("duplicate connections must be rejected")
	}
	navigator, err := service.Join(explorer.Code)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Join(explorer.Code); !errors.Is(err, usecase.ErrRoomFull) {
		t.Fatal("only one navigator may join")
	}
	navigatorViews, leaveNavigator, err := service.Subscribe(explorer.Code, navigator.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveNavigator()
	navigation := <-navigatorViews
	if navigation.Friend == nil || navigation.Cells[1][7] == nil || len(navigation.Lanterns) != 2 {
		t.Fatal("navigator should receive the full map and entities")
	}
	if ready := <-explorerViews; ready.Phase != "playing" {
		t.Fatal("both connections should start play")
	}
	now := time.Now()
	if _, err := service.Act(explorer.Code, navigator.Token, protocol.ClientMessage{Type: "move", Direction: "right"}, now); !errors.Is(err, domain.ErrWrongRole) {
		t.Fatal("navigator must not move the explorer")
	}
	if _, err := service.Act(explorer.Code, navigator.Token, protocol.ClientMessage{Type: "hint", Direction: "right"}, now); err != nil {
		t.Fatal(err)
	}
	if hint := <-explorerViews; hint.Hint != "Go right" {
		t.Fatal("directions should reach the explorer")
	}
	moved, err := service.Act(explorer.Code, explorer.Token, protocol.ClientMessage{Type: "move", Direction: "right"}, now)
	if err != nil || !moved {
		t.Fatal("explorer should move onto the first lantern")
	}
	lit := <-explorerViews
	if lit.LightRemainingMS != 20000 || lit.VisionRadius != 3 || lit.LanternsCollected != 1 || lit.Cells[1][5] == nil || lit.Friend != nil {
		t.Fatal("light expands vision but must not leak the distant friend")
	}
	service.Tick(now.Add(5*time.Second), 5*time.Second)
	if active := <-explorerViews; active.LightRemainingMS != 15000 {
		t.Fatal("server tick should advance light")
	}
	leaveNavigator()
	if paused := <-explorerViews; paused.Phase != "paused" {
		t.Fatal("disconnect should pause the room")
	}
	service.Tick(now.Add(10*time.Second), 5*time.Second)
	if _, err := service.Act(explorer.Code, explorer.Token, protocol.ClientMessage{Type: "move", Direction: "right"}, now.Add(time.Second)); !errors.Is(err, domain.ErrRoomPaused) {
		t.Fatal("paused rooms must reject movement")
	}
	navigatorViews, leaveAgain, err := service.Subscribe(explorer.Code, navigator.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer leaveAgain()
	if reconnected := <-navigatorViews; reconnected.LightRemainingMS != 15000 || reconnected.Moves != 1 || reconnected.Phase != "playing" {
		t.Fatal("reconnect must preserve the game and paused light")
	}
	service.Tick(now.Add(25*time.Second), 15*time.Second)
	if dark := <-explorerViews; dark.LightRemainingMS != 0 || dark.VisionRadius != 1 || dark.Cells[1][5] != nil {
		t.Fatal("expiry must hide terrain beyond radius one")
	}
	for step := 0; step < 5; step++ {
		if _, err := service.Act(explorer.Code, explorer.Token, protocol.ClientMessage{Type: "move", Direction: "right"}, now.Add(time.Duration(26+step)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	rescued := <-explorerViews
	if rescued.Phase != "won" || rescued.Friend == nil {
		t.Fatal("finding the friend should win for both players")
	}
	for _, row := range rescued.Cells {
		for _, cell := range row {
			if cell == nil {
				t.Fatal("winning should reveal the whole map")
			}
		}
	}
	if _, err := service.Act(explorer.Code, explorer.Token, protocol.ClientMessage{Type: "restart"}, now.Add(31*time.Second)); err != nil {
		t.Fatal(err)
	}
	reset := <-explorerViews
	if reset.Round != 2 || reset.Moves != 0 || reset.LanternsCollected != 0 || reset.LightRemainingMS != 0 || reset.LanternTotal != 2 {
		t.Fatal("restart should synchronize a fresh round")
	}
}

func TestRoomsExpireAfterBothPlayersLeave(t *testing.T) {
	service := usecase.NewRoomService(memory.NewRoomRepository())
	session, err := service.Create()
	if err != nil {
		t.Fatal(err)
	}
	_, leave, err := service.Subscribe(session.Code, session.Token)
	if err != nil {
		t.Fatal(err)
	}
	leave()
	leave()
	service.Tick(time.Now().Add(21*time.Minute), time.Second)
	if _, err := service.Join(session.Code); !errors.Is(err, usecase.ErrRoomNotFound) {
		t.Fatal("abandoned rooms should expire")
	}
}
