package usecase

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/protocol"
)

var (
	ErrRoomNotFound     = errors.New("room not found or expired")
	ErrRoomFull         = errors.New("this room already has a navigator")
	ErrInvalidSession   = errors.New("invalid room session")
	ErrAlreadyConnected = errors.New("this player is already connected in another window")
	ErrRoomCapacity     = errors.New("the server is full; please try again later")
	ErrTooFast          = errors.New("please wait before sending another action")
)

// RoomRecord is application session state, mutated only inside repository operations.
type RoomRecord struct {
	Code         string
	Room         *domain.Room
	Tokens       map[domain.Role]string
	Streams      map[domain.Role]chan protocol.RoomView
	LastAction   map[domain.Role]time.Time
	LastActive   time.Time
	Round        int
	Revision     int
	VoiceStreams map[domain.Role]chan protocol.VoiceMessage
	VoiceCallID  string
}

type RoomRepository interface {
	Insert(*RoomRecord) error
	Update(string, func(*RoomRecord) error) error
	Sweep(func(*RoomRecord) bool)
}

type RoomService struct{ repository RoomRepository }

func NewRoomService(repository RoomRepository) *RoomService {
	return &RoomService{repository: repository}
}

func (s *RoomService) Create() (protocol.Session, error) {
	room, err := domain.NewRoom()
	if err != nil {
		return protocol.Session{}, err
	}
	code, token := rand.Text()[:8], rand.Text()
	record := &RoomRecord{Code: code, Room: room, Tokens: map[domain.Role]string{domain.Explorer: token}, Streams: map[domain.Role]chan protocol.RoomView{}, LastAction: map[domain.Role]time.Time{}, LastActive: time.Now(), Round: 1}
	if err := s.repository.Insert(record); err != nil {
		return protocol.Session{}, err
	}
	return protocol.Session{Code: code, Token: token, Role: domain.Explorer}, nil
}

func normalizeCode(code string) string { return strings.ToUpper(strings.TrimSpace(code)) }

func (s *RoomService) Join(code string) (protocol.Session, error) {
	var session protocol.Session
	err := s.repository.Update(normalizeCode(code), func(r *RoomRecord) error {
		if r.Tokens[domain.Navigator] != "" {
			return ErrRoomFull
		}
		token := rand.Text()
		r.Tokens[domain.Navigator], r.LastActive = token, time.Now()
		session = protocol.Session{Code: r.Code, Token: token, Role: domain.Navigator}
		return nil
	})
	return session, err
}

func roleFor(r *RoomRecord, token string) (domain.Role, error) {
	if token == "" {
		return "", ErrInvalidSession
	}
	for role, expected := range r.Tokens {
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1 {
			return role, nil
		}
	}
	return "", ErrInvalidSession
}

func (s *RoomService) Subscribe(code, token string) (<-chan protocol.RoomView, func(), error) {
	var stream chan protocol.RoomView
	var role domain.Role
	code = normalizeCode(code)
	err := s.repository.Update(code, func(r *RoomRecord) error {
		var err error
		role, err = roleFor(r, token)
		if err != nil {
			return err
		}
		if r.Streams[role] != nil {
			return ErrAlreadyConnected
		}
		stream = make(chan protocol.RoomView, 1)
		r.Streams[role] = stream
		r.Room.Connected[role], r.LastActive = true, time.Now()
		publish(r)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return stream, func() {
		_ = s.repository.Update(code, func(r *RoomRecord) error {
			if r.Streams[role] != stream {
				return nil
			}
			delete(r.Streams, role)
			r.Room.Connected[role], r.LastActive = false, time.Now()
			close(stream)
			publish(r)
			return nil
		})
	}, nil
}

func (s *RoomService) Act(code, token string, message protocol.ClientMessage, now time.Time) (bool, error) {
	moved := false
	err := s.repository.Update(normalizeCode(code), func(r *RoomRecord) error {
		role, err := roleFor(r, token)
		if err != nil {
			return err
		}
		if !r.Room.Connected[role] {
			return ErrInvalidSession
		}
		if now.Sub(r.LastAction[role]) < 180*time.Millisecond {
			return ErrTooFast
		}
		switch message.Type {
		case "move":
			moved, err = r.Room.Move(role, message.Direction)
		case "hint":
			err = r.Room.Guide(role, message.Direction)
		case "restart":
			err = r.Room.Restart(role)
			if err == nil {
				r.Round++
			}
		default:
			return errors.New("unsupported room message")
		}
		if err != nil {
			return err
		}
		r.LastAction[role], r.LastActive = now, now
		publish(r)
		return nil
	})
	return moved, err
}

func roomView(r *RoomRecord, role domain.Role) protocol.RoomView {
	game := r.Room.Game
	view := protocol.RoomView{Code: r.Code, Role: role, Round: r.Round, Revision: r.Revision, Phase: "waiting", Player: protocol.Position{X: game.Player.Position.X, Y: game.Player.Position.Y}, Moves: game.Moves, Hint: r.Room.Hint, Lanterns: []protocol.Position{}, LanternsCollected: game.LanternsCollected, LanternTotal: game.LanternsCollected + len(game.Lanterns), LightRemainingMS: game.LightRemaining.Milliseconds(), VisionRadius: game.VisionRadius()}
	if game.Status == domain.GameStatusWon {
		view.Phase = "won"
	} else if r.Room.Ready() {
		view.Phase = "playing"
	} else if r.Tokens[domain.Navigator] != "" {
		view.Phase = "paused"
	}
	view.Cells = make([][]*bool, len(game.Maze.Cells))
	for y, row := range game.Maze.Cells {
		view.Cells[y] = make([]*bool, len(row))
		for x, walkable := range row {
			if r.Room.Visible(role, domain.Position{X: x, Y: y}) {
				cell := walkable
				view.Cells[y][x] = &cell
			}
		}
	}
	if r.Room.Visible(role, game.Goal) {
		view.Friend = &protocol.Position{X: game.Goal.X, Y: game.Goal.Y}
	}
	for _, lantern := range game.Lanterns {
		if r.Room.Visible(role, lantern) {
			view.Lanterns = append(view.Lanterns, protocol.Position{X: lantern.X, Y: lantern.Y})
		}
	}
	return view
}

func publish(r *RoomRecord) {
	r.Revision++
	for role, stream := range r.Streams {
		view := roomView(r, role)
		select {
		case stream <- view:
		default:
			select {
			case <-stream:
			default:
			}
			select {
			case stream <- view:
			default:
			}
		}
	}
}

func (s *RoomService) Tick(now time.Time, elapsed time.Duration) {
	s.repository.Sweep(func(r *RoomRecord) bool {
		if len(r.Streams) == 0 && len(r.VoiceStreams) == 0 && now.Sub(r.LastActive) > 20*time.Minute {
			return true
		}
		if r.Room.Ready() && r.Room.Game.LightRemaining > 0 && r.Room.Game.Status == domain.GameStatusPlaying {
			r.Room.Game.AdvanceLight(elapsed)
			publish(r)
		}
		return false
	})
}

func (s *RoomService) Run(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	previous := time.Now()
	for {
		select {
		case now := <-ticker.C:
			s.Tick(now, now.Sub(previous))
			previous = now
		case <-ctx.Done():
			s.repository.Sweep(func(r *RoomRecord) bool {
				for _, stream := range r.Streams {
					close(stream)
				}
				for _, stream := range r.VoiceStreams {
					close(stream)
				}
				return true
			})
			return
		}
	}
}
