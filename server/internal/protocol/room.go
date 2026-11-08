package protocol

import "github.com/putriindaht/leftnt/server/internal/domain"

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Session struct {
	Code  string      `json:"code"`
	Token string      `json:"token"`
	Role  domain.Role `json:"role"`
}

// Null cells are unknown. No hidden terrain or entities are sent to an explorer.
type RoomView struct {
	Code              string      `json:"code"`
	Role              domain.Role `json:"role"`
	Phase             string      `json:"phase"`
	Round             int         `json:"round"`
	Revision          int         `json:"revision"`
	Cells             [][]*bool   `json:"cells"`
	Player            Position    `json:"player"`
	Friend            *Position   `json:"friend"`
	Lanterns          []Position  `json:"lanterns"`
	LanternsCollected int         `json:"lanternsCollected"`
	LanternTotal      int         `json:"lanternTotal"`
	LightRemainingMS  int64       `json:"lightRemainingMs"`
	VisionRadius      int         `json:"visionRadius"`
	Moves             int         `json:"moves"`
	Hint              string      `json:"hint"`
}

type ClientMessage struct {
	Type      string `json:"type"`
	Token     string `json:"token,omitempty"`
	Direction string `json:"direction,omitempty"`
}

type ServerMessage struct {
	Type  string    `json:"type"`
	State *RoomView `json:"state,omitempty"`
	Error string    `json:"error,omitempty"`
	Moved *bool     `json:"moved,omitempty"`
}
