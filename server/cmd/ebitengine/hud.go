package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const hudHeight = 76

func restartBounds() image.Rectangle {
	width := mazeWidth * tileSize
	top := mazeHeight * tileSize
	return image.Rect(width-96, top+52, width-8, top+70)
}

func (g *Game) restartButtonHovered() bool {
	x, y := ebiten.CursorPosition()
	return g.pointerPosition(x, y).In(restartBounds())
}

func (g *Game) restartRequested() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && g.restartButtonHovered() {
		return true
	}
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		if g.pointerPosition(x, y).In(restartBounds()) {
			return true
		}
	}
	return false
}

func drawHUDText(screen *ebiten.Image, message string, x, y int, shade color.Color) {
	drawGameText(screen, message, x, y, shade, hudFont, 11)
}

func drawGameText(screen *ebiten.Image, message string, x, y int, shade color.Color, face text.Face, lineSpacing float64) {
	options := &text.DrawOptions{}
	options.GeoM.Translate(float64(x), float64(y))
	options.Filter = ebiten.FilterNearest
	options.ColorScale.ScaleWithColor(shade)
	options.LineSpacing = lineSpacing
	text.Draw(screen, message, face, options)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	top := mazeHeight * tileSize
	width := mazeWidth * tileSize
	background := color.RGBA{R: 35, G: 31, B: 45, A: 255}
	foreground := color.RGBA{R: 255, G: 241, B: 211, A: 255}
	muted := color.RGBA{R: 198, G: 181, B: 161, A: 255}
	gold := color.RGBA{R: 239, G: 193, B: 109, A: 255}
	ink := color.RGBA{R: 51, G: 39, B: 24, A: 255}
	vector.FillRect(screen, 0, float32(top), float32(width), hudHeight, background, false)
	vector.FillRect(screen, 0, float32(top), float32(width), 1, gold, false)

	title := "Find your lost friend"
	titleColor := color.Color(foreground)
	if g.friendFound() {
		title = "Friend found!"
		titleColor = gold
	}
	drawGameText(screen, title, 8, top+7, titleColor, hudTitleFont, 14)

	total := g.state.LanternsCollected + len(g.state.Lanterns)
	status := fmt.Sprintf("Moves: %d | Lanterns: %d/%d", g.state.Moves, g.state.LanternsCollected, total)
	drawHUDText(screen, status, 8, top+21, muted)
	light := "off"
	lightColor := color.Color(muted)
	if g.state.LightRemaining > 0 {
		seconds := (g.state.LightRemaining + time.Second - 1) / time.Second
		light = fmt.Sprintf("%ds", seconds)
		lightColor = gold
	}
	drawHUDText(screen, fmt.Sprintf("Light: %s | Sight: %d", light, g.state.VisionRadius()), 8, top+33, lightColor)

	controls := "Arrows/WASD: move\nT: switch walls"
	if g.friendFound() {
		controls = "Together again. Nice rescue!\nPlay another maze?"
	}
	drawHUDText(screen, controls, 8, top+45, foreground)

	button := restartBounds()
	buttonColor := gold
	if g.restartHovered {
		buttonColor = color.RGBA{R: 255, G: 224, B: 160, A: 255}
	}
	vector.FillRect(screen, float32(button.Min.X), float32(button.Min.Y), float32(button.Dx()), float32(button.Dy()), buttonColor, false)
	label := "Restart [R]"
	drawHUDText(screen, label, button.Min.X+(button.Dx()-int(text.Advance(label, hudFont)))/2, button.Min.Y+4, ink)
}
