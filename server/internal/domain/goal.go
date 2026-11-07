package domain

import (
	"errors"
	"math/rand/v2"
)

func (m Maze) RandomGoal(start Position) (Position, error) {
	if !m.IsWalkable(start) {
		return Position{}, errors.New("start position must be walkable")
	}

	distances := m.distancesFrom(start)

	maxDistance := 0

	for _, distance := range distances {
		if distance > maxDistance {
			maxDistance = distance
		}
	}

	var candidates []Position

	for position, distance := range distances {
		if position == start {
			continue
		}

		if distance*10 >= maxDistance*7 {
			candidates = append(candidates, position)
		}
	}

	if len(candidates) == 0 {
		return Position{}, errors.New("no available path cell for the goal")
	}

	index := rand.IntN(len(candidates))
	return candidates[index], nil
}

func (m Maze) distancesFrom(start Position) map[Position]int {
	distances := map[Position]int{
		start: 0,
	}

	queue := []Position{start}

	directions := []Position{
		{X: 0, Y: -1},
		{X: 0, Y: 1},
		{X: -1, Y: 0},
		{X: 1, Y: 0},
	}

	for head := 0; head < len(queue); head++ {
		current := queue[head]

		for _, direction := range directions {
			next := Position{
				X: current.X + direction.X,
				Y: current.Y + direction.Y,
			}

			if !m.IsWalkable(next) {
				continue
			}

			if _, visited := distances[next]; visited {
				continue
			}

			distances[next] = distances[current] + 1
			queue = append(queue, next)
		}
	}

	return distances
}
