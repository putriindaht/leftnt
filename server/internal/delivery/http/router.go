package httpdelivery

import (
	"github.com/putriindaht/leftnt/server/internal/usecase"
	"net/http"
)

func NewRouter(checker HealthChecker, games *usecase.GameService, rooms ...*RoomHandler) *http.ServeMux {
	mux := http.NewServeMux()
	healthHandler := NewHealthHandler(checker)
	mux.HandleFunc("GET /health", healthHandler.Check)
	gameHandler := NewGameHandler(games)
	mux.HandleFunc("POST /games", gameHandler.Create)
	mux.HandleFunc("GET /games/{id}", gameHandler.Get)
	mux.HandleFunc("POST /games/{id}/moves", gameHandler.Move)
	if len(rooms) > 0 && rooms[0] != nil {
		mux.HandleFunc("POST /rooms", rooms[0].Create)
		mux.HandleFunc("POST /rooms/{code}/join", rooms[0].Join)
		mux.HandleFunc("GET /rooms/{code}/ws", rooms[0].Socket)
		mux.HandleFunc("GET /rooms/{code}/voice", rooms[0].Voice)
	}
	return mux
}
