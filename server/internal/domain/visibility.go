package domain

func (m Maze) CanSee(from, target Position) bool {
	if !m.IsWalkable(from) {
		return false
	}

	if target.Y < 0 || target.Y >= len(m.Cells) {
		return false
	}

	if target.X < 0 || target.X >= len(m.Cells[target.Y]) {
		return false
	}

	x, y := from.X, from.Y
	dx := target.X - from.X
	dy := target.Y - from.Y

	stepX, stepY := 1, 1

	if dx < 0 {
		stepX = -1
		dx = -dx
	}

	if dy < 0 {
		stepY = -1
		dy = -dy
	}

	crossedX, crossedY := 0, 0

	for x != target.X || y != target.Y {
		horizontal := (1 + 2*crossedX) * dy
		vertical := (1 + 2*crossedY) * dx

		switch {
		case horizontal < vertical:
			x += stepX
			crossedX++

		case horizontal > vertical:
			y += stepY
			crossedY++

		default:
			sideX := Position{X: x + stepX, Y: y}
			sideY := Position{X: x, Y: y + stepY}

			if !m.IsWalkable(sideX) || !m.IsWalkable(sideY) {
				return false
			}

			x += stepX
			y += stepY
			crossedX++
			crossedY++
		}

		if x == target.X && y == target.Y {
			return true
		}

		if !m.IsWalkable(Position{X: x, Y: y}) {
			return false
		}
	}

	return true
}
