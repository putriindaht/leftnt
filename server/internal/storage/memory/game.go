package memory

import (
	"crypto/rand"
	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/usecase"
	"sync"
)

type GameRepository struct {
	mu    sync.RWMutex
	games map[string]domain.Game
}

func NewGameRepository() *GameRepository {
	return &GameRepository{games: make(map[string]domain.Game)}
}

func (r *GameRepository) Create(game domain.Game) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := rand.Text()
	for {
		if _, exists := r.games[id]; !exists {
			break
		}
		id = rand.Text()
	}
	r.games[id] = snapshot(game)
	return id, nil
}

func (r *GameRepository) Find(id string) (domain.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	game, found := r.games[id]
	if !found {
		return domain.Game{}, usecase.ErrGameNotFound
	}
	return snapshot(game), nil
}

func (r *GameRepository) Update(id string, apply func(*domain.Game) error) (domain.Game, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	game, found := r.games[id]
	if !found {
		return domain.Game{}, usecase.ErrGameNotFound
	}
	game = snapshot(game)
	if err := apply(&game); err != nil {
		return domain.Game{}, err
	}
	r.games[id] = snapshot(game)
	return game, nil
}

// Copy rows so callers cannot mutate stored state through a shared slice.
func snapshot(game domain.Game) domain.Game {
	rows := make([][]bool, len(game.Maze.Cells))
	for y, row := range game.Maze.Cells {
		rows[y] = append([]bool(nil), row...)
	}
	game.Maze.Cells = rows
	return game
}
