package domain

type Position struct {
	X int
	Y int
}

type Maze struct {
	Cells [][]bool
}

func (m Maze) IsWalkable(position Position) bool {
	if position.Y < 0 || position.Y >= len(m.Cells) {
		return false
	}

	row := m.Cells[position.Y]

	if position.X < 0 || position.X >= len(row) {
		return false
	}

	return row[position.X]
}
