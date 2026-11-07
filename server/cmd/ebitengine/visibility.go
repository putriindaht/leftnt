package main

import (
	"math"

	"github.com/putriindaht/leftnt/server/internal/domain"
)

// Terrain, pickups, and the friend all share the current light radius and sight rules.
func withinSight(maze domain.Maze, target domain.Position, centerX, centerY float64, radius int) bool {
	if !tileVisible(target.X, target.Y, centerX, centerY, radius, false) {
		return false
	}
	observer := domain.Position{X: int(math.Round(centerX)), Y: int(math.Round(centerY))}
	return maze.CanSee(observer, target)
}
