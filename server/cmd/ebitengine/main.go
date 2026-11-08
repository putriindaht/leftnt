package main

import (
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	roomclient "github.com/putriindaht/leftnt/server/internal/client"
	"github.com/putriindaht/leftnt/server/internal/domain"
)

const (
	mazeWidth   = 15
	mazeHeight  = 11
	tileSize    = 16
	windowScale = 3
)

type Game struct {
	multiplayer      *multiplayerState
	state            *domain.SoloGame
	dirtTile         *ebiten.Image
	woodTile         *ebiten.Image
	hedgeTile        *ebiten.Image
	useHedge         bool
	playerSprites    []*ebiten.Image
	walk             walkAnimation
	queuedDirection  string
	winAnnounced     bool
	feedback         wallFeedback
	blockedDirection string
	restartHovered   bool
	pointerTransform ebiten.GeoM
	lanternSprite    *ebiten.Image
	friendSprite     *ebiten.Image
}

func (g *Game) Update() error {
	if g.multiplayer != nil {
		return g.updateMultiplayer()
	}
	g.feedback.advance()
	g.restartHovered = g.restartButtonHovered()
	if g.restartRequested() {
		return g.restart()
	}
	g.state.AdvanceLight(time.Second / time.Duration(ebiten.TPS()))

	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.useHedge = !g.useHedge
	}

	// Remember a quick turn pressed while the previous step is finishing.
	if direction := movementDirection(inpututil.IsKeyJustPressed); direction != "" {
		g.queuedDirection = direction
		g.blockedDirection = ""
	}
	if movementDirection(ebiten.IsKeyPressed) == "" && g.queuedDirection == "" {
		g.blockedDirection = ""
	}

	g.walk.advance()
	if g.walk.moving() {
		return nil
	}

	if g.state.Status != domain.GameStatusPlaying {
		g.queuedDirection = ""
		if !g.winAnnounced && g.friendFound() {
			ebiten.SetWindowTitle("LEFTN'T - Friend found!")
			g.winAnnounced = true
		}
		return nil
	}

	direction := g.queuedDirection
	g.queuedDirection = ""
	if direction == "" {
		direction = movementDirection(ebiten.IsKeyPressed)
	}
	if direction == "" {
		return nil
	}
	return g.startMove(direction)
}

func movementDirection(pressed func(ebiten.Key) bool) string {
	switch {
	case pressed(ebiten.KeyArrowUp) || pressed(ebiten.KeyW):
		return "up"
	case pressed(ebiten.KeyArrowDown) || pressed(ebiten.KeyS):
		return "down"
	case pressed(ebiten.KeyArrowLeft) || pressed(ebiten.KeyA):
		return "left"
	case pressed(ebiten.KeyArrowRight) || pressed(ebiten.KeyD):
		return "right"
	default:
		return ""
	}
}

func (g *Game) startMove(direction string) error {
	from := g.state.Player.Position
	moved, err := g.state.Move(direction)
	if err != nil {
		return err
	}
	if moved {
		g.blockedDirection = ""
		g.walk.begin(from, g.state.Player.Position)
	} else if g.blockedDirection != direction {
		g.blockedDirection = direction
		g.feedback.bump()
	}
	return nil
}

func (g *Game) friendFound() bool {
	if g.multiplayer != nil {
		return g.multiplayer.view != nil && g.multiplayer.view.Phase == "won" && !g.walk.moving()
	}
	return g.state.Status == domain.GameStatusWon && !g.walk.moving()
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.multiplayer != nil {
		g.drawMultiplayer(screen)
		return
	}
	renderX, renderY := g.walk.position()
	found := g.friendFound()
	radius := g.state.VisionRadius()

	screen.Fill(color.RGBA{
		R: 24, G: 26, B: 42, A: 255,
	})

	for y, row := range g.state.Maze.Cells {
		for x, walkable := range row {
			target := domain.Position{X: x, Y: y}
			if !found && !withinSight(g.state.Maze, target, renderX, renderY, radius) {
				continue
			}

			tile := g.woodTile

			if g.useHedge {
				tile = g.hedgeTile
			}

			if walkable {
				tile = g.dirtTile
			}

			options := &ebiten.DrawImageOptions{}
			options.Filter = ebiten.FilterNearest
			if !found {
				if g.state.LightRemaining > 0 {
					// Lantern light brings back warm, bright terrain colors.
					options.ColorScale.Scale(1.12, 1.05, 0.92, 1)
				} else {
					// Keep nearby paths readable in the cool ambient darkness.
					options.ColorScale.Scale(0.45, 0.47, 0.60, 1)
				}
			}

			options.GeoM.Translate(
				float64(x*tileSize),
				float64(y*tileSize),
			)

			screen.DrawImage(tile, options)
		}
	}

	for _, lantern := range g.state.Lanterns {
		if found || withinSight(g.state.Maze, lantern, renderX, renderY, radius) {
			drawSprite(screen, g.lanternSprite, float64(lantern.X*tileSize), float64(lantern.Y*tileSize))
		}
	}
	if found || withinSight(g.state.Maze, g.state.Goal, renderX, renderY, radius) {
		friendX := float64(g.state.Goal.X * tileSize)
		if found {
			friendX += 5
		}
		drawSprite(screen, g.friendSprite, friendX, float64(g.state.Goal.Y*tileSize))
	}

	// Draw the player last so the character stays visible over the goal.
	playerOptions := &ebiten.DrawImageOptions{}
	playerOptions.Filter = ebiten.FilterNearest
	playerX := math.Round(renderX * tileSize)
	if found {
		playerX -= 5
	}

	playerOptions.GeoM.Translate(
		playerX,
		math.Round(renderY*tileSize),
	)

	screen.DrawImage(g.playerSprites[g.walk.frame()], playerOptions)
	g.drawWallBubble(screen, renderX, renderY)
	g.drawHUD(screen)

}

func drawSprite(screen, sprite *ebiten.Image, x, y float64) {
	options := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	options.GeoM.Translate(x, y)
	screen.DrawImage(sprite, options)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.multiplayer != nil {
		return mazeWidth * tileSize, mazeHeight*tileSize + multiplayerHUDHeight
	}
	return mazeWidth * tileSize, mazeHeight*tileSize + hudHeight
}

func main() {
	game := &Game{
		dirtTile:      newDirtTile(),
		woodTile:      newWoodTile(),
		hedgeTile:     newHedgeTile(),
		useHedge:      true,
		playerSprites: newPlayerSprites(),
		lanternSprite: newLanternSprite(),
		friendSprite:  newFriendSprite(),
	}
	height := mazeHeight*tileSize + hudHeight
	if endpoint := os.Getenv("LEFTNT_WS_URL"); endpoint != "" {
		token := os.Getenv("LEFTNT_ROOM_TOKEN")
		if token == "" {
			log.Fatal("LEFTNT_ROOM_TOKEN is required for multiplayer")
		}
		network := roomclient.NewRoomClient(endpoint, token)
		defer network.Close()
		game.multiplayer = &multiplayerState{client: network}
		game.walk = newWalkAnimation(domain.Position{X: 1, Y: 1})
		height = mazeHeight*tileSize + multiplayerHUDHeight
	} else {
		state, err := domain.NewSoloGame(mazeWidth, mazeHeight, domain.DefaultSoloConfig())
		if err != nil {
			log.Fatal(err)
		}
		game.state = state
		game.walk = newWalkAnimation(state.Player.Position)
	}

	ebiten.SetWindowSize(
		mazeWidth*tileSize*windowScale,
		height*windowScale,
	)
	ebiten.SetWindowTitle("LEFTN'T - Ebitengine")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
