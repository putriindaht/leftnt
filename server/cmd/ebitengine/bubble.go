package main

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const bubbleColumns = 28
const bubbleLineSpacing = 11

type speechBubbleLayout struct {
	bounds image.Rectangle
	tailX  int
	below  bool
}

func layoutWallBubble(message string, playerX, playerY float64) speechBubbleLayout {
	lines := strings.Split(wrapMessage(message, bubbleColumns), "\n")
	longest := 0.0
	for _, line := range lines {
		longest = max(longest, text.Advance(line, bubbleFont))
	}
	width := int(math.Ceil(longest)) + 12
	height := len(lines)*bubbleLineSpacing + 10
	boardWidth, boardHeight := mazeWidth*tileSize, mazeHeight*tileSize
	anchorX := int(math.Round((playerX + 0.5) * tileSize))
	playerTop := int(math.Round(playerY * tileSize))
	x := max(4, min(anchorX-width/2, boardWidth-width-4))
	y := playerTop - height - 6
	below := y < 4
	if below {
		y = playerTop + tileSize + 6
	}
	y = max(4, min(y, boardHeight-height-4))
	tailX := max(x+6, min(anchorX, x+width-7))
	return speechBubbleLayout{bounds: image.Rect(x, y, x+width, y+height), tailX: tailX, below: below}
}

func (g *Game) drawWallBubble(screen *ebiten.Image, playerX, playerY float64) {
	if g.feedback.remaining <= 0 || g.friendFound() {
		return
	}
	layout := layoutWallBubble(g.feedback.message, playerX, playerY)
	box := layout.bounds
	ink := color.RGBA{R: 68, G: 44, B: 37, A: 255}
	paper := color.RGBA{R: 255, G: 241, B: 211, A: 255}
	vector.FillRect(screen, float32(box.Min.X), float32(box.Min.Y), float32(box.Dx()), float32(box.Dy()), ink, false)
	vector.FillRect(screen, float32(box.Min.X+1), float32(box.Min.Y+1), float32(box.Dx()-2), float32(box.Dy()-2), paper, false)

	// Use small pixel steps for a speech tail pointing toward the explorer.
	if layout.below {
		vector.FillRect(screen, float32(layout.tailX-3), float32(box.Min.Y-2), 7, 2, ink, false)
		vector.FillRect(screen, float32(layout.tailX-1), float32(box.Min.Y-4), 3, 2, ink, false)
		vector.FillRect(screen, float32(layout.tailX-1), float32(box.Min.Y-1), 3, 2, paper, false)
	} else {
		vector.FillRect(screen, float32(layout.tailX-3), float32(box.Max.Y), 7, 2, ink, false)
		vector.FillRect(screen, float32(layout.tailX-1), float32(box.Max.Y+2), 3, 2, ink, false)
		vector.FillRect(screen, float32(layout.tailX-1), float32(box.Max.Y-1), 3, 2, paper, false)
	}
	drawGameText(screen, wrapMessage(g.feedback.message, bubbleColumns), box.Min.X+6, box.Min.Y+5, ink, bubbleFont, bubbleLineSpacing)
}
