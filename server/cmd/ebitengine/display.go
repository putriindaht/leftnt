package main

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Keep every game pixel the same physical size, including on high-DPI screens.
// Small windows crop at 1x rather than shrinking and blurring the text.
func (g *Game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, defaultTransform ebiten.GeoM) {
	output := screen.Bounds()
	input := offscreen.Bounds()
	deviceScale := ebiten.Monitor().DeviceScaleFactor()
	fit := min(float64(output.Dx())/float64(input.Dx()), float64(output.Dy())/float64(input.Dy()))
	// Snap in window pixels too, so Retina displays do not choose a 2.5x scale.
	windowScale := max(1, math.Floor(fit/deviceScale))
	scale := math.Round(windowScale * deviceScale)
	options := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	options.GeoM.Scale(scale, scale)
	options.GeoM.Translate(
		math.Floor((float64(output.Dx())-float64(input.Dx())*scale)/2/deviceScale)*deviceScale,
		math.Floor((float64(output.Dy())-float64(input.Dy())*scale)/2/deviceScale)*deviceScale,
	)
	screen.DrawImage(offscreen, options)
	// Ebitengine reports input in its default layout. Match our letterboxed view.
	inverse := options.GeoM
	inverse.Invert()
	g.pointerTransform = defaultTransform
	g.pointerTransform.Concat(inverse)
}

func (g *Game) pointerPosition(x, y int) image.Point {
	px, py := g.pointerTransform.Apply(float64(x), float64(y))
	return image.Pt(int(math.Floor(px)), int(math.Floor(py)))
}
