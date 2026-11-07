package usecase

import (
	"errors"
	"github.com/putriindaht/leftnt/server/internal/domain"
)

var ErrGameNotFound = errors.New("game not found")

// GameRepository provides snapshots and atomic updates of stored games.
type GameRepository interface {
	Create(game domain.Game) (string, error)
	Find(id string) (domain.Game, error)
	Update(id string, apply func(*domain.Game) error) (domain.Game, error)
}

type GameService struct{ repository GameRepository }

func NewGameService(repository GameRepository) *GameService {
	return &GameService{repository: repository}
}

func (s *GameService) Create(width, height int) (string, domain.Game, error) {
	game, err := domain.NewGame(width, height)
	if err != nil {
		return "", domain.Game{}, err
	}
	id, err := s.repository.Create(*game)
	return id, *game, err
}

func (s *GameService) Get(id string) (domain.Game, error) {
	return s.repository.Find(id)
}

func (s *GameService) Move(id, direction string) (domain.Game, bool, error) {
	moved := false
	game, err := s.repository.Update(id, func(game *domain.Game) error {
		var err error
		moved, err = game.Move(direction)
		return err
	})
	return game, moved, err
}
