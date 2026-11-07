package domain

import (
	"errors"
	"math/rand/v2"
	"time"
)

// SoloConfig controls optional light pickups for the find-your-friend mode.
type SoloConfig struct {
	LanternCount        int
	LanternDuration     time.Duration
	BaseVisionRadius    int
	LanternVisionRadius int
}

func DefaultSoloConfig() SoloConfig {
	return SoloConfig{
		LanternCount:        2,
		LanternDuration:     20 * time.Second,
		BaseVisionRadius:    1,
		LanternVisionRadius: 3,
	}
}

// SoloGame adds timed light to the shared maze rules. Goal is the lost friend.
type SoloGame struct {
	*Game
	Lanterns          []Position
	LanternsCollected int
	LightRemaining    time.Duration
	config            SoloConfig
}

func NewSoloGame(width, height int, config SoloConfig) (*SoloGame, error) {
	if config.LanternCount < 0 || config.LanternDuration <= 0 || config.BaseVisionRadius < 1 || config.LanternVisionRadius <= config.BaseVisionRadius {
		return nil, errors.New("solo settings require a nonnegative lantern count, positive duration, and a lantern radius greater than the base radius")
	}
	game, err := NewGame(width, height)
	if err != nil {
		return nil, err
	}
	return &SoloGame{
		Game:     game,
		Lanterns: game.Maze.randomLanterns(game.Player.Position, game.Goal, config.LanternCount),
		config:   config,
	}, nil
}

func (g *SoloGame) Move(direction string) (bool, error) {
	moved, err := g.Game.Move(direction)
	if err != nil || !moved {
		return moved, err
	}
	for i, position := range g.Lanterns {
		if position == g.Player.Position {
			g.Lanterns = append(g.Lanterns[:i], g.Lanterns[i+1:]...)
			g.LanternsCollected++
			// Refilling replaces the remaining duration; it does not stack.
			g.LightRemaining = g.config.LanternDuration
			break
		}
	}
	return true, nil
}

// AdvanceLight receives elapsed game time; the domain does not read a clock.
func (g *SoloGame) AdvanceLight(elapsed time.Duration) {
	if elapsed <= 0 || g.Status != GameStatusPlaying {
		return
	}
	g.LightRemaining = max(0, g.LightRemaining-elapsed)
}

func (g *SoloGame) VisionRadius() int {
	if g.LightRemaining > 0 {
		return g.config.LanternVisionRadius
	}
	return g.config.BaseVisionRadius
}

// Place one pickup in each section of the route to the friend. Separate
// sections spread pickups out, while random positions vary between runs.
// Start and goal are excluded, and small mazes use only available route cells.
func (m Maze) randomLanterns(start, goal Position, count int) []Position {
	if count == 0 {
		return nil
	}
	distances := m.distancesFrom(start)
	distance, reachable := distances[goal]
	if !reachable || distance < 2 {
		return nil
	}
	route := make([]Position, distance+1)
	current := goal
	directions := []Position{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}
	for step := distance; step >= 0; step-- {
		route[step] = current
		for _, direction := range directions {
			previous := Position{X: current.X + direction.X, Y: current.Y + direction.Y}
			if previousDistance, ok := distances[previous]; ok && previousDistance == step-1 {
				current = previous
				break
			}
		}
	}
	available := route[1 : len(route)-1]
	count = min(count, len(available))
	lanterns := make([]Position, 0, count)
	for i := 0; i < count; i++ {
		low := i * len(available) / count
		high := (i + 1) * len(available) / count
		// Sample the middle half of each section to avoid adjacent pickups
		// across section boundaries when the route is long enough.
		inset := (high - low) / 4
		low += inset
		high -= inset
		lanterns = append(lanterns, available[low+rand.IntN(high-low)])
	}
	return lanterns
}
