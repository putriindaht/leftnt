package domain

import "errors"

type Role string

const (
	Explorer  Role = "explorer"
	Navigator Role = "navigator"
)

var (
	ErrWrongRole  = errors.New("only the explorer can move or start a new maze")
	ErrRoomPaused = errors.New("both players must be connected to play")
)

type Room struct {
	Game      *SoloGame
	Connected map[Role]bool
	Hint      string
}

func NewRoom() (*Room, error) {
	game, err := NewSoloGame(15, 11, DefaultSoloConfig())
	if err != nil {
		return nil, err
	}
	return &Room{Game: game, Connected: map[Role]bool{}}, nil
}

func (r *Room) Ready() bool { return r.Connected[Explorer] && r.Connected[Navigator] }

func (r *Room) Move(role Role, direction string) (bool, error) {
	if role != Explorer {
		return false, ErrWrongRole
	}
	if !r.Ready() {
		return false, ErrRoomPaused
	}
	return r.Game.Move(direction)
}

func (r *Room) Guide(role Role, direction string) error {
	if role != Navigator {
		return errors.New("only the navigator can send directions")
	}
	if !r.Ready() {
		return ErrRoomPaused
	}
	if r.Game.Status == GameStatusWon {
		return ErrGameFinished
	}
	switch direction {
	case "up", "down", "left", "right":
		r.Hint = "Go " + direction
	default:
		return ErrInvalidDirection
	}
	return nil
}

func (r *Room) Restart(role Role) error {
	if role != Explorer {
		return ErrWrongRole
	}
	if !r.Ready() {
		return ErrRoomPaused
	}
	game, err := NewSoloGame(15, 11, DefaultSoloConfig())
	if err != nil {
		return err
	}
	r.Game, r.Hint = game, ""
	return nil
}

func (r *Room) Visible(role Role, position Position) bool {
	if role == Navigator || r.Game.Status == GameStatusWon {
		return true
	}
	from := r.Game.Player.Position
	dx, dy := position.X-from.X, position.Y-from.Y
	radius := r.Game.VisionRadius()
	return dx*dx+dy*dy <= radius*radius && r.Game.Maze.CanSee(from, position)
}
