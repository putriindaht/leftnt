package main

import "github.com/putriindaht/leftnt/server/internal/domain"

const walkDuration = 12

// walkAnimation changes only the presentation; domain positions stay on the grid.
type walkAnimation struct {
	from    domain.Position
	target  domain.Position
	elapsed int
	steps   int
}

func newWalkAnimation(start domain.Position) walkAnimation {
	return walkAnimation{from: start, target: start, elapsed: walkDuration}
}

func (w *walkAnimation) begin(from, target domain.Position) {
	w.from = from
	w.target = target
	w.elapsed = 0
	w.steps++
}

func (w *walkAnimation) advance() {
	if w.moving() {
		w.elapsed++
	}
}

func (w walkAnimation) moving() bool {
	return w.elapsed < walkDuration
}

func (w walkAnimation) position() (float64, float64) {
	progress := float64(w.elapsed) / float64(walkDuration)
	x := float64(w.from.X) + float64(w.target.X-w.from.X)*progress
	y := float64(w.from.Y) + float64(w.target.Y-w.from.Y)*progress
	return x, y
}

func (w walkAnimation) frame() int {
	if !w.moving() || w.elapsed < 2 || w.elapsed >= walkDuration-2 {
		return 0
	}
	if w.steps%2 == 1 {
		return 1
	}
	return 2
}

func tileVisible(x, y int, centerX, centerY float64, radius int, revealAll bool) bool {
	if revealAll {
		return true
	}
	dx := float64(x) - centerX
	dy := float64(y) - centerY
	return dx*dx+dy*dy <= float64(radius*radius)
}
