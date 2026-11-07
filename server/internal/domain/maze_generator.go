package domain

import (
	"errors"
	"math/rand/v2"
)

func NewMaze(width, height int) (Maze, error) {
	if width < 3 || height < 3 {
		return Maze{}, errors.New("maze dimensions must be at least 3")
	}

	if width%2 == 0 || height%2 == 0 {
		return Maze{}, errors.New("maze dimensions must be odd")
	}

	cells := make([][]bool, height)

	for y := range cells {
		cells[y] = make([]bool, width)
	}

	cells[1][1] = true

	return Maze{Cells: cells}, nil
}

func GenerateMaze(width, height int) (Maze, error) {
	maze, err := NewMaze(width, height)
	if err != nil {
		return Maze{}, err
	}

	start := Position{X: 1, Y: 1}
	stack := []Position{start}

	directions := []Position{
		{X: 0, Y: -1},
		{X: 0, Y: 1},
		{X: -1, Y: 0},
		{X: 1, Y: 0},
	}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		var candidates []Position

		for _, direction := range directions {
			target := Position{
				X: current.X + direction.X*2,
				Y: current.Y + direction.Y*2,
			}

			if target.X <= 0 || target.X >= width-1 ||
				target.Y <= 0 || target.Y >= height-1 {
				continue
			}

			if maze.Cells[target.Y][target.X] {
				continue
			}

			candidates = append(candidates, target)
		}

		if len(candidates) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}

		target := candidates[rand.IntN(len(candidates))]

		wall := Position{
			X: (current.X + target.X) / 2,
			Y: (current.Y + target.Y) / 2,
		}

		maze.Cells[wall.Y][wall.X] = true
		maze.Cells[target.Y][target.X] = true

		stack = append(stack, target)
	}

	return maze, nil
}
