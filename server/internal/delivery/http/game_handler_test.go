package httpdelivery

import (
	"encoding/json"
	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/storage/memory"
	"github.com/putriindaht/leftnt/server/internal/usecase"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGameAPI(t *testing.T) {
	repository := memory.NewGameRepository()
	router := NewRouter(usecase.NewHealthService(), usecase.NewGameService(repository))
	call := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		response := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		if response.Code != status {
			t.Fatalf("%s %s: expected %d, got %d: %s", method, path, status, response.Code, response.Body.String())
		}
		return response
	}
	created := call(http.MethodPost, "/games", `{"width":15,"height":11}`, http.StatusCreated)
	var game gameResponse
	if err := json.Unmarshal(created.Body.Bytes(), &game); err != nil {
		t.Fatal(err)
	}
	if game.ID == "" || game.Status != domain.GameStatusPlaying || game.Maze.Width != 15 || game.Maze.Height != 11 {
		t.Fatalf("invalid game response: %+v", game)
	}
	if created.Header().Get("Location") != "/games/"+game.ID {
		t.Fatal("missing game location")
	}
	fetched := call(http.MethodGet, "/games/"+game.ID, "", http.StatusOK)
	if fetched.Body.String() != created.Body.String() {
		t.Fatal("GET should return the stored game")
	}
	call(http.MethodGet, "/games/missing", "", http.StatusNotFound)
	call(http.MethodPost, "/games/missing/moves", `{"direction":"right"}`, http.StatusNotFound)
	for _, body := range []string{`null`, `[]`, `{`, `{} {}`, `{"width":14}`, `{"width":3,"height":3}`, `{"width":1001}`, `{"unknown":true}`, `{"width":"15"}`} {
		call(http.MethodPost, "/games", body, http.StatusBadRequest)
	}
	call(http.MethodPost, "/games/"+game.ID+"/moves", `{"direction":"teleport"}`, http.StatusBadRequest)
	call(http.MethodPost, "/games/"+game.ID+"/moves", `{}`, http.StatusBadRequest)
	call(http.MethodPost, "/games/"+game.ID+"/moves", `{"direction":"up","player":{"x":9}}`, http.StatusBadRequest)

	// Store a predictable corridor to test collision, persistence, and winning.
	id, err := repository.Create(domain.Game{Maze: domain.Maze{Cells: [][]bool{{false, false, false, false, false}, {false, true, true, true, false}, {false, false, false, false, false}}}, Player: domain.Player{Position: domain.Position{X: 1, Y: 1}}, Goal: domain.Position{X: 3, Y: 1}, Status: domain.GameStatusPlaying})
	if err != nil {
		t.Fatal(err)
	}
	blocked := call(http.MethodPost, "/games/"+id+"/moves", `{"direction":"up"}`, http.StatusOK)
	var blockedGame gameResponse
	if err := json.Unmarshal(blocked.Body.Bytes(), &blockedGame); err != nil {
		t.Fatal(err)
	}
	if blockedGame.Moved == nil || *blockedGame.Moved || blockedGame.Moves != 0 {
		t.Fatal("blocked moves should return moved=false without incrementing moves")
	}
	call(http.MethodPost, "/games/"+id+"/moves", `{"direction":"right"}`, http.StatusOK)
	won := call(http.MethodPost, "/games/"+id+"/moves", `{"direction":"right"}`, http.StatusOK)
	var wonGame gameResponse
	if err := json.Unmarshal(won.Body.Bytes(), &wonGame); err != nil {
		t.Fatal(err)
	}
	if wonGame.Status != domain.GameStatusWon || wonGame.Moves != 2 || wonGame.Player != wonGame.Goal {
		t.Fatal("goal should be reached in two moves")
	}
	stored := call(http.MethodGet, "/games/"+id, "", http.StatusOK)
	var storedGame gameResponse
	if err := json.Unmarshal(stored.Body.Bytes(), &storedGame); err != nil {
		t.Fatal(err)
	}
	if storedGame.Status != domain.GameStatusWon || storedGame.Player != wonGame.Player || storedGame.Moves != 2 {
		t.Fatal("GET should return the persisted winning state")
	}
	call(http.MethodPost, "/games/"+id+"/moves", `{"direction":"left"}`, http.StatusConflict)
	call(http.MethodGet, "/health", "", http.StatusOK)
}
