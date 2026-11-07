package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func newPlayerSprite() *ebiten.Image {
	sprite := ebiten.NewImage(tileSize, tileSize)

	palette := map[rune]color.RGBA{
		'h': {R: 78, G: 49, B: 35, A: 255},
		'r': {R: 221, G: 119, B: 142, A: 255},
		's': {R: 239, G: 190, B: 146, A: 255},
		'e': {R: 30, G: 29, B: 40, A: 255},
		'b': {R: 69, G: 104, B: 168, A: 255},
		'l': {R: 114, G: 154, B: 209, A: 255},
		'd': {R: 57, G: 72, B: 102, A: 255},
		't': {R: 166, G: 104, B: 53, A: 255},
		'p': {R: 78, G: 64, B: 60, A: 255},
		'k': {R: 45, G: 35, B: 35, A: 255},
	}

	pixels := []string{
		"................",
		"......hhhh......",
		".....hhhhhh.....",
		"..hhrhsssshrhh..",
		"..hh.sesses.hh..",
		"..hh.hssssh.hh..",
		"...h.bbllbb.h...",
		"....sbbllbbtt...",
		"....sbbllbbtt...",
		"....sbbbbbbtt...",
		".....dddddd.....",
		".....pp..pp.....",
		".....pp..pp.....",
		"....kkk..kkk....",
		"................",
		"................",
	}

	for y, row := range pixels {
		for x, pixel := range row {
			shade, exists := palette[pixel]

			if exists {
				sprite.Set(x, y, shade)
			}
		}
	}

	return sprite
}

func newLanternSprite() *ebiten.Image {
	sprite := ebiten.NewImage(tileSize, tileSize)

	palette := map[rune]color.RGBA{
		'b': {R: 91, G: 58, B: 32, A: 255},
		'h': {R: 170, G: 112, B: 49, A: 255},
		'g': {R: 235, G: 162, B: 53, A: 255},
		'y': {R: 255, G: 208, B: 79, A: 255},
		'l': {R: 255, G: 245, B: 190, A: 255},
		'o': {R: 246, G: 128, B: 45, A: 255},
	}

	pixels := []string{
		"................",
		".......bb.......",
		"......b..b......",
		"......b..b......",
		".....bbbbbb.....",
		"....bhhhhhhb....",
		"....bgyyyygb....",
		"....bgyllygb....",
		"....bgyllygb....",
		"....bgyooygb....",
		"....bgyyyygb....",
		".....bbbbbb.....",
		"....bhhhhhhb....",
		".....bbbbbb.....",
		"................",
		"................",
	}

	for y, row := range pixels {
		for x, pixel := range row {
			shade, exists := palette[pixel]
			if exists {
				sprite.Set(x, y, shade)
			}
		}
	}

	return sprite
}

// The lost friend has short hair and a red shirt, distinct from the blue explorer.
func newFriendSprite() *ebiten.Image {
	sprite := ebiten.NewImage(tileSize, tileSize)
	palette := map[rune]color.RGBA{
		'h': {R: 66, G: 42, B: 32, A: 255},
		's': {R: 226, G: 171, B: 127, A: 255},
		'e': {R: 30, G: 29, B: 40, A: 255},
		'r': {R: 183, G: 53, B: 55, A: 255},
		'l': {R: 236, G: 101, B: 82, A: 255},
		'd': {R: 125, G: 39, B: 43, A: 255},
		'p': {R: 132, G: 112, B: 77, A: 255},
		'k': {R: 45, G: 35, B: 35, A: 255},
	}
	pixels := []string{
		"................",
		".....hhhhhh.....",
		"....hhhhhhhh....",
		"....hhsssshh....",
		"....hsessesh....",
		".....ssssss.....",
		"..ss..rrrr......",
		"..ss.rllrrr.....",
		"...ssrllrrrs....",
		".....rrrrrrs....",
		".....dddddd.....",
		".....pp..pp.....",
		".....pp..pp.....",
		"....kkk..kkk....",
		"................",
		"................",
	}
	for y, row := range pixels {
		for x, pixel := range row {
			if shade, exists := palette[pixel]; exists {
				sprite.Set(x, y, shade)
			}
		}
	}
	return sprite
}

// newPlayerSprites caches idle, left-step, and right-step images once at startup.
func newPlayerSprites() []*ebiten.Image {
	idle := newPlayerSprite()
	return []*ebiten.Image{
		idle,
		newWalkingFrame(idle, 12, 14),
		newWalkingFrame(idle, 14, 12),
	}
}

func newWalkingFrame(idle *ebiten.Image, leftFootY, rightFootY int) *ebiten.Image {
	sprite := ebiten.NewImage(tileSize, tileSize)
	sprite.DrawImage(idle, nil)

	// Keep the head, pigtails, jacket, and bag; redraw just the legs and boots.
	for y := 11; y < tileSize; y++ {
		for x := 0; x < tileSize; x++ {
			sprite.Set(x, y, color.RGBA{})
		}
	}
	pantsColor := color.RGBA{R: 78, G: 64, B: 60, A: 255}
	bootColor := color.RGBA{R: 45, G: 35, B: 35, A: 255}
	legs := []struct{ pantsX, bootX, footY int }{
		{5, 4, leftFootY},
		{9, 9, rightFootY},
	}
	for _, leg := range legs {
		for y := 11; y < leg.footY; y++ {
			sprite.Set(leg.pantsX, y, pantsColor)
			sprite.Set(leg.pantsX+1, y, pantsColor)
		}
		for x := leg.bootX; x < leg.bootX+3; x++ {
			sprite.Set(x, leg.footY, bootColor)
		}
	}
	return sprite
}
