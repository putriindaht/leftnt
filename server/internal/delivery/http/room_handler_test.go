package httpdelivery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/protocol"
	"github.com/putriindaht/leftnt/server/internal/storage/memory"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

func roomTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	service := usecase.NewRoomService(memory.NewRoomRepository())
	server := httptest.NewServer(NewRouter(usecase.NewHealthService(), usecase.NewGameService(memory.NewGameRepository()), NewRoomHandler(service, nil)))
	t.Cleanup(server.Close)
	return server
}

func roomHTTP(t *testing.T, server *httptest.Server, path string, want int) protocol.Session {
	t.Helper()
	response, err := http.Post(server.URL+path, "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != want {
		t.Fatalf("HTTP %s: got %d, want %d", path, response.StatusCode, want)
	}
	var session protocol.Session
	if want < 300 {
		if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
			t.Fatal(err)
		}
	}
	return session
}

func roomSocket(t *testing.T, server *httptest.Server, session protocol.Session) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/rooms/"+session.Code+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	if err := wsjson.Write(ctx, conn, protocol.ClientMessage{Type: "authenticate", Token: session.Token}); err != nil {
		t.Fatal(err)
	}
	return conn
}

func readRoomMessage(t *testing.T, conn *websocket.Conn, want func(protocol.ServerMessage) bool) protocol.ServerMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for {
		var message protocol.ServerMessage
		if err := wsjson.Read(ctx, conn, &message); err != nil {
			t.Fatal(err)
		}
		if want(message) {
			return message
		}
	}
}

func TestWebSocketTwoPlayersDirectionsMovementAndReconnect(t *testing.T) {
	server := roomTestServer(t)
	explorer := roomHTTP(t, server, "/rooms", http.StatusCreated)
	if explorer.Role != domain.Explorer || len(explorer.Code) != 8 || explorer.Token == "" {
		t.Fatal("create should return an explorer session")
	}
	a := roomSocket(t, server, explorer)
	readRoomMessage(t, a, func(m protocol.ServerMessage) bool { return m.State != nil && m.State.Phase == "waiting" })
	navigator := roomHTTP(t, server, "/rooms/"+explorer.Code+"/join", http.StatusOK)
	roomHTTP(t, server, "/rooms/"+explorer.Code+"/join", http.StatusConflict)
	b := roomSocket(t, server, navigator)
	full := readRoomMessage(t, b, func(m protocol.ServerMessage) bool { return m.State != nil && m.State.Phase == "playing" }).State
	private := readRoomMessage(t, a, func(m protocol.ServerMessage) bool { return m.State != nil && m.State.Phase == "playing" }).State
	if full.Friend == nil || private.Friend != nil {
		t.Fatal("friend visibility must be role-specific")
	}
	for _, row := range full.Cells {
		for _, cell := range row {
			if cell == nil {
				t.Fatal("navigator map must be complete")
			}
		}
	}
	if err := wsjson.Write(context.Background(), b, protocol.ClientMessage{Type: "move", Direction: "right"}); err != nil {
		t.Fatal(err)
	}
	readRoomMessage(t, b, func(m protocol.ServerMessage) bool { return strings.Contains(m.Error, "only the explorer") })
	if err := wsjson.Write(context.Background(), b, protocol.ClientMessage{Type: "hint", Direction: "right"}); err != nil {
		t.Fatal(err)
	}
	readRoomMessage(t, a, func(m protocol.ServerMessage) bool { return m.State != nil && m.State.Hint == "Go right" })
	direction := "down"
	if *full.Cells[1][2] {
		direction = "right"
	}
	if err := wsjson.Write(context.Background(), a, protocol.ClientMessage{Type: "move", Direction: direction}); err != nil {
		t.Fatal(err)
	}
	moved := readRoomMessage(t, a, func(m protocol.ServerMessage) bool { return m.State != nil && m.State.Moves == 1 }).State
	readRoomMessage(t, b, func(m protocol.ServerMessage) bool { return m.State != nil && m.State.Player == moved.Player })
	b.CloseNow()
	readRoomMessage(t, a, func(m protocol.ServerMessage) bool { return m.State != nil && m.State.Phase == "paused" })
	reconnected := roomSocket(t, server, navigator)
	resumed := readRoomMessage(t, reconnected, func(m protocol.ServerMessage) bool { return m.State != nil && m.State.Phase == "playing" }).State
	if resumed.Moves != 1 || resumed.Player != moved.Player {
		t.Fatal("reconnect must resume the same game")
	}
}

func TestWebSocketRejectsInvalidTokenAndForeignOrigin(t *testing.T) {
	server := roomTestServer(t)
	session := roomHTTP(t, server, "/rooms", http.StatusCreated)
	session.Token = "wrong"
	conn := roomSocket(t, server, session)
	rejected := readRoomMessage(t, conn, func(m protocol.ServerMessage) bool { return m.Error != "" })
	if rejected.State != nil || !strings.Contains(rejected.Error, "invalid room session") {
		t.Fatal("unauthorized clients must receive no game state")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	foreign, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/rooms/"+session.Code+"/ws", &websocket.DialOptions{HTTPHeader: http.Header{"Origin": []string{"https://unrelated.example"}}})
	if foreign != nil {
		foreign.CloseNow()
	}
	if err == nil || response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatal("unapproved browser origins must be rejected")
	}
	roomHTTP(t, server, "/rooms/UNKNOWN2/join", http.StatusNotFound)
}
