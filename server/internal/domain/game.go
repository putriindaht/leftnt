package domain

import "errors"

type GameStatus string

const (
	GameStatusPlaying GameStatus = "playing"
	GameStatusWon     GameStatus = "won"
)

var (
	ErrInvalidDimensions = errors.New("maze dimensions must be odd numbers between 3 and 51, with at least one dimension greater than 3")
	ErrInvalidDirection  = errors.New("direction must be up, down, left, or right")
	ErrGameFinished      = errors.New("game has already finished")
)

type Game struct {
	Maze   Maze
	Player Player
	Goal   Position
	Status GameStatus
	Moves  int
}

func NewGame(width, height int) (*Game, error) {
	if width < 3 || height < 3 || width > 51 || height > 51 || width%2 == 0 || height%2 == 0 || (width == 3 && height == 3) {
		return nil, ErrInvalidDimensions
	}
	maze, err := GenerateMaze(width, height)
	if err != nil {
		return nil, err
	}
	player := Player{Position: Position{X: 1, Y: 1}}
	goal, err := maze.RandomGoal(player.Position)
	if err != nil {
		return nil, err
	}
	return &Game{Maze: maze, Player: player, Goal: goal, Status: GameStatusPlaying}, nil
}

func (g *Game) Move(direction string) (bool, error) {
	if g.Status != GameStatusPlaying {
		return false, ErrGameFinished
	}
	switch direction {
	case "up", "down", "left", "right":
	default:
		return false, ErrInvalidDirection
	}
	if !g.Player.Move(direction, g.Maze) {
		return false, nil
	}
	g.Moves++
	if g.Player.Position == g.Goal {
		g.Status = GameStatusWon
	}
	return true, nil
}
