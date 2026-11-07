package httpdelivery

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/usecase"
)

type GameHandler struct{ service *usecase.GameService }

func NewGameHandler(service *usecase.GameService) *GameHandler { return &GameHandler{service: service} }

type positionResponse struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type gameResponse struct {
	ID     string            `json:"id"`
	Maze   mazeResponse      `json:"maze"`
	Player positionResponse  `json:"player"`
	Goal   positionResponse  `json:"goal"`
	Status domain.GameStatus `json:"status"`
	Moves  int               `json:"moves"`
	Moved  *bool             `json:"moved,omitempty"`
}

type mazeResponse struct {
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Cells  [][]bool `json:"cells"`
}

func gameView(id string, game domain.Game) gameResponse {
	return gameResponse{
		ID: id,
		Maze: mazeResponse{
			Width:  len(game.Maze.Cells[0]),
			Height: len(game.Maze.Cells),
			Cells:  game.Maze.Cells,
		},
		Player: positionResponse{X: game.Player.Position.X, Y: game.Player.Position.Y},
		Goal:   positionResponse{X: game.Goal.X, Y: game.Goal.Y},
		Status: game.Status,
		Moves:  game.Moves,
	}
}

func (h *GameHandler) Create(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}{Width: 15, Height: 11}
	if !decodeJSON(w, r, &request) {
		return
	}
	id, game, err := h.service.Create(request.Width, request.Height)
	if err != nil {
		writeGameError(w, err)
		return
	}
	w.Header().Set("Location", "/games/"+id)
	writeJSON(w, http.StatusCreated, gameView(id, game))
}

func (h *GameHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	game, err := h.service.Get(id)
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, gameView(id, game))
}

func (h *GameHandler) Move(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Direction string `json:"direction"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	id := r.PathValue("id")
	game, moved, err := h.service.Move(id, request.Direction)
	if err != nil {
		writeGameError(w, err)
		return
	}
	response := gameView(id, game)
	response.Moved = &moved
	writeJSON(w, http.StatusOK, response)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request must contain a valid JSON object with supported fields"})
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request must contain exactly one JSON object"})
		return false
	}
	strict := json.NewDecoder(bytes.NewReader(raw))
	strict.DisallowUnknownFields()
	if len(raw) == 0 || raw[0] != '{' {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request must be a JSON object"})
		return false
	}
	if err := strict.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request contains invalid or unsupported fields"})
		return false
	}
	return true
}

func writeGameError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "failed to process game request"
	switch {
	case errors.Is(err, usecase.ErrGameNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, domain.ErrInvalidDimensions), errors.Is(err, domain.ErrInvalidDirection):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, domain.ErrGameFinished):
		status = http.StatusConflict
		message = err.Error()
	default:
		log.Printf("game request failed: %v", err)
	}
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write game response: %v", err)
	}
}
