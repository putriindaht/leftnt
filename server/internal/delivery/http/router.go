package httpdelivery

import (
	"github.com/putriindaht/leftnt/server/internal/usecase"
	"net/http"
)

func NewRouter(checker HealthChecker, games *usecase.GameService) *http.ServeMux {
	mux := http.NewServeMux()
	healthHandler := NewHealthHandler(checker)
	mux.HandleFunc("GET /health", healthHandler.Check)
	gameHandler := NewGameHandler(games)
	mux.HandleFunc("POST /games", gameHandler.Create)
	mux.HandleFunc("GET /games/{id}", gameHandler.Get)
	mux.HandleFunc("POST /games/{id}/moves", gameHandler.Move)
	return mux
}
