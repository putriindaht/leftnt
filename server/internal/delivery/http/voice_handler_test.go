package httpdelivery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/putriindaht/leftnt/server/internal/protocol"
	"github.com/putriindaht/leftnt/server/internal/storage/memory"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

func voiceSocket(t *testing.T, server *httptest.Server, session protocol.Session) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/rooms/"+session.Code+"/voice", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	if err := wsjson.Write(ctx, conn, protocol.VoiceMessage{Type: "authenticate", Token: session.Token}); err != nil {
		t.Fatal(err)
	}
	return conn
}
func readVoice(t *testing.T, conn *websocket.Conn) protocol.VoiceMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var message protocol.VoiceMessage
	if err := wsjson.Read(ctx, conn, &message); err != nil {
		t.Fatal(err)
	}
	return message
}

func TestVoiceWebSocketReadyRelayAndLeave(t *testing.T) {
	service := usecase.NewRoomService(memory.NewRoomRepository())
	ice := []protocol.ICEServer{{URLs: []string{"stun:stun.example.test:3478"}}}
	server := httptest.NewServer(NewRouter(usecase.NewHealthService(), usecase.NewGameService(memory.NewGameRepository()), NewRoomHandler(service, nil, ice)))
	defer server.Close()
	explorer := roomHTTP(t, server, "/rooms", http.StatusCreated)
	navigator := roomHTTP(t, server, "/rooms/"+explorer.Code+"/join", http.StatusOK)
	a := voiceSocket(t, server, explorer)
	waiting := readVoice(t, a)
	if waiting.PeerReady || len(waiting.ICEServers) != 1 {
		t.Fatal("authenticated ready should include configured ICE settings")
	}
	b := voiceSocket(t, server, navigator)
	readyA, readyB := readVoice(t, a), readVoice(t, b)
	if !readyA.PeerReady || readyA.CallID != readyB.CallID {
		t.Fatal("both voice sockets should share a call")
	}
	offer := protocol.VoiceMessage{Type: "description", CallID: readyA.CallID, Description: &protocol.SessionDescription{Type: "offer", SDP: "v=0\r\n"}}
	if err := wsjson.Write(context.Background(), a, offer); err != nil {
		t.Fatal(err)
	}
	if received := readVoice(t, b); received.Description == nil || received.Description.SDP != offer.Description.SDP {
		t.Fatal("offer was not relayed through the socket")
	}
	b.CloseNow()
	if readVoice(t, a).Type != "peer-left" {
		t.Fatal("socket disconnect should notify the partner")
	}
}

func TestVoiceWebSocketRejectsTokenAndOrigin(t *testing.T) {
	server := roomTestServer(t)
	session := roomHTTP(t, server, "/rooms", http.StatusCreated)
	session.Token = "wrong"
	invalid := voiceSocket(t, server, session)
	if message := readVoice(t, invalid); message.Type != "error" || message.Description != nil {
		t.Fatal("invalid sessions must not receive connection details")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/rooms/"+session.Code+"/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Origin": []string{"https://unrelated.example"}}})
	if conn != nil {
		conn.CloseNow()
	}
	if err == nil || response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatal("voice must reject unapproved browser origins")
	}
}
