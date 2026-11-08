package memory

import (
	"sync"

	"github.com/putriindaht/leftnt/server/internal/usecase"
)

type RoomRepository struct {
	mu    sync.Mutex
	rooms map[string]*usecase.RoomRecord
}

func NewRoomRepository() *RoomRepository {
	return &RoomRepository{rooms: map[string]*usecase.RoomRecord{}}
}

func (r *RoomRepository) Insert(room *usecase.RoomRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.rooms) >= 256 {
		return usecase.ErrRoomCapacity
	}
	if _, exists := r.rooms[room.Code]; exists {
		return usecase.ErrRoomCapacity
	}
	r.rooms[room.Code] = room
	return nil
}

func (r *RoomRepository) Update(code string, apply func(*usecase.RoomRecord) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[code]
	if !ok {
		return usecase.ErrRoomNotFound
	}
	return apply(room)
}

func (r *RoomRepository) Sweep(remove func(*usecase.RoomRecord) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for code, room := range r.rooms {
		if remove(room) {
			delete(r.rooms, code)
		}
	}
}
