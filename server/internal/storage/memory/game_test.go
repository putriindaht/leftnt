package memory

import (
	"errors"
	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/usecase"
	"sync"
	"testing"
)

func TestRepositorySnapshotsAndAtomicUpdates(t *testing.T) {
	repository := NewGameRepository()
	original := domain.Game{Maze: domain.Maze{Cells: [][]bool{{true}}}}
	id, err := repository.Create(original)
	if err != nil {
		t.Fatal(err)
	}
	original.Maze.Cells[0][0] = false
	snapshot, err := repository.Find(id)
	if err != nil || !snapshot.Maze.Cells[0][0] {
		t.Fatal("create should copy the original grid")
	}
	snapshot.Maze.Cells[0][0] = false
	stored, _ := repository.Find(id)
	if !stored.Maze.Cells[0][0] {
		t.Fatal("read snapshots must not modify stored state")
	}
	rollback := errors.New("reject update")
	_, err = repository.Update(id, func(game *domain.Game) error { game.Maze.Cells[0][0] = false; return rollback })
	if !errors.Is(err, rollback) {
		t.Fatal("expected rejected update")
	}
	stored, _ = repository.Find(id)
	if !stored.Maze.Cells[0][0] {
		t.Fatal("rejected updates must not change stored state")
	}
	var workers sync.WaitGroup
	for i := 0; i < 100; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := repository.Update(id, func(game *domain.Game) error { game.Moves++; return nil })
			if err != nil {
				t.Error(err)
			}
		}()
	}
	workers.Wait()
	stored, _ = repository.Find(id)
	if stored.Moves != 100 {
		t.Fatalf("expected 100 atomic updates, got %d", stored.Moves)
	}
	if _, err := repository.Find("missing"); !errors.Is(err, usecase.ErrGameNotFound) {
		t.Fatal("expected missing game error")
	}
}
