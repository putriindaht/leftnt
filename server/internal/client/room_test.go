package client_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	roomclient "github.com/putriindaht/leftnt/server/internal/client"
	httpdelivery "github.com/putriindaht/leftnt/server/internal/delivery/http"
	"github.com/putriindaht/leftnt/server/internal/protocol"
	"github.com/putriindaht/leftnt/server/internal/storage/memory"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

func nextView(t *testing.T, client *roomclient.RoomClient, matches func(protocol.RoomView) bool) protocol.RoomView {
	t.Helper()
	timeout := time.NewTimer(4 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case event := <-client.Events:
			if event.Message.State != nil && matches(*event.Message.State) {
				return *event.Message.State
			}
		case <-timeout.C:
			t.Fatal("timed out waiting for the native room client")
			return protocol.RoomView{}
		}
	}
}

func TestNativeRoomClientAuthenticatesSynchronizesAndResumes(t *testing.T) {
	service := usecase.NewRoomService(memory.NewRoomRepository())
	server := httptest.NewServer(httpdelivery.NewRouter(usecase.NewHealthService(), usecase.NewGameService(memory.NewGameRepository()), httpdelivery.NewRoomHandler(service, nil)))
	defer server.Close()
	explorer, err := service.Create()
	if err != nil {
		t.Fatal(err)
	}
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http") + "/rooms/" + explorer.Code + "/ws"
	a := roomclient.NewRoomClient(endpoint, explorer.Token)
	defer a.Close()
	nextView(t, a, func(view protocol.RoomView) bool { return view.Phase == "waiting" })
	navigator, err := service.Join(explorer.Code)
	if err != nil {
		t.Fatal(err)
	}
	b := roomclient.NewRoomClient(endpoint, navigator.Token)
	defer b.Close()
	full := nextView(t, b, func(view protocol.RoomView) bool { return view.Phase == "playing" })
	nextView(t, a, func(view protocol.RoomView) bool { return view.Phase == "playing" })
	direction := "down"
	if *full.Cells[1][2] {
		direction = "right"
	}
	if !a.Send(protocol.ClientMessage{Type: "move", Direction: direction}) {
		t.Fatal("client should accept a movement command")
	}
	moved := nextView(t, a, func(view protocol.RoomView) bool { return view.Moves == 1 })
	nextView(t, b, func(view protocol.RoomView) bool { return view.Player == moved.Player && view.Moves == 1 })
	b.Close()
	nextView(t, a, func(view protocol.RoomView) bool { return view.Phase == "paused" })
	resumed := roomclient.NewRoomClient(endpoint, navigator.Token)
	defer resumed.Close()
	preserved := nextView(t, resumed, func(view protocol.RoomView) bool { return view.Phase == "playing" })
	if preserved.Player != moved.Player || preserved.Moves != 1 {
		t.Fatal("the native client should resume the existing game")
	}
}
