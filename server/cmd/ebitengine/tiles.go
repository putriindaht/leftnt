package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func newDirtTile() *ebiten.Image {
	tile := ebiten.NewImage(tileSize, tileSize)

	baseColor := color.RGBA{
		R: 177, G: 151, B: 128, A: 255,
	}

	detailColor := color.RGBA{
		R: 145, G: 117, B: 94, A: 255,
	}

	tile.Fill(baseColor)

	tile.Set(3, 4, detailColor)
	tile.Set(11, 2, detailColor)
	tile.Set(7, 10, detailColor)
	tile.Set(13, 13, detailColor)

	return tile
}

func newWoodTile() *ebiten.Image {
	tile := ebiten.NewImage(tileSize, tileSize)

	shadowColor := color.RGBA{
		R: 72, G: 43, B: 29, A: 255,
	}

	plankColor := color.RGBA{
		R: 139, G: 91, B: 55, A: 255,
	}

	highlightColor := color.RGBA{
		R: 173, G: 119, B: 73, A: 255,
	}

	grainColor := color.RGBA{
		R: 103, G: 64, B: 39, A: 255,
	}

	tile.Fill(shadowColor)

	// Draw two wooden planks.
	vector.FillRect(tile, 1, 1, 6, 14, plankColor, false)
	vector.FillRect(tile, 8, 1, 7, 14, plankColor, false)

	// Add a light edge to each plank.
	vector.FillRect(tile, 1, 1, 1, 14, highlightColor, false)
	vector.FillRect(tile, 8, 1, 1, 14, highlightColor, false)

	// Add darker wood grain.
	vector.FillRect(tile, 4, 4, 1, 6, grainColor, false)
	vector.FillRect(tile, 11, 7, 1, 5, grainColor, false)

	return tile
}

func newHedgeTile() *ebiten.Image {
	tile := ebiten.NewImage(tileSize, tileSize)

	shadowColor := color.RGBA{
		R: 35, G: 61, B: 38, A: 255,
	}

	leafColor := color.RGBA{
		R: 67, G: 104, B: 55, A: 255,
	}

	highlightColor := color.RGBA{
		R: 110, G: 143, B: 74, A: 255,
	}

	tile.Fill(shadowColor)

	// Draw the main foliage.
	vector.FillRect(tile, 1, 2, 14, 12, leafColor, false)

	// Add clusters of leaves catching the light.
	vector.FillRect(tile, 2, 2, 5, 3, highlightColor, false)
	vector.FillRect(tile, 9, 4, 4, 2, highlightColor, false)

	// Add darker areas between the leaves.
	vector.FillRect(tile, 3, 10, 4, 3, shadowColor, false)
	vector.FillRect(tile, 10, 11, 3, 2, shadowColor, false)

	return tile
}
