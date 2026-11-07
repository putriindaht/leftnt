package domain

type Player struct {
	Position Position
}

func (p *Player) Move(direction string, maze Maze) bool {
	next := p.Position

	switch direction {
	case "up":
		next.Y--
	case "down":
		next.Y++
	case "left":
		next.X--
	case "right":
		next.X++
	default:
		return false
	}

	if !maze.IsWalkable(next) {
		return false
	}

	p.Position = next
	return true
}
