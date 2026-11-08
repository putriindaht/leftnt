package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	roomclient "github.com/putriindaht/leftnt/server/internal/client"
	"github.com/putriindaht/leftnt/server/internal/domain"
	"github.com/putriindaht/leftnt/server/internal/protocol"
)

const multiplayerHUDHeight = 120

type multiplayerState struct {
	client       *roomclient.RoomClient
	view         *protocol.RoomView
	connected    bool
	connection   string
	errorMessage string
	errorTicks   int
	cooldown     int
}

func (g *Game) updateMultiplayer() error {
	network := g.multiplayer
	g.feedback.advance()
	g.walk.advance()
	if network.cooldown > 0 {
		network.cooldown--
	}
	if network.errorTicks > 0 {
		network.errorTicks--
	}
	for {
		select {
		case event := <-network.client.Events:
			if event.Connection != "" {
				network.connected, network.connection = false, event.Connection
				g.queuedDirection = ""
			}
			message := event.Message
			if message.Error != "" {
				network.errorMessage, network.errorTicks = message.Error, 240
			}
			if message.Moved != nil {
				if *message.Moved {
					g.blockedDirection = ""
				} else if g.blockedDirection == "" {
					g.blockedDirection = "wall"
					g.feedback.bump()
				}
			}
			if message.State != nil {
				view := message.State
				target := domain.Position{X: view.Player.X, Y: view.Player.Y}
				if network.view == nil || network.view.Round != view.Round {
					g.walk = newWalkAnimation(target)
					g.feedback = wallFeedback{}
					g.queuedDirection = ""
				} else if network.view.Player != view.Player {
					from := domain.Position{X: network.view.Player.X, Y: network.view.Player.Y}
					if abs(target.X-from.X)+abs(target.Y-from.Y) == 1 {
						g.walk.begin(from, target)
					} else {
						g.walk = newWalkAnimation(target)
					}
				}
				network.view, network.connected = view, true
			}
		default:
			goto drained
		}
	}
drained:
	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.useHedge = !g.useHedge
	}
	if network.view == nil || !network.connected {
		return nil
	}
	direction := movementDirection(inpututil.IsKeyJustPressed)
	if direction != "" {
		g.blockedDirection = ""
		g.queuedDirection = direction
	}
	if movementDirection(ebiten.IsKeyPressed) == "" && g.queuedDirection == "" {
		g.blockedDirection = ""
	}
	if clicked := g.multiplayerClick(); clicked != "" {
		if clicked == "restart" {
			g.sendRoomCommand("restart", "")
		} else {
			g.queuedDirection = clicked
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.sendRoomCommand("restart", "")
	}
	if network.view.Phase != "playing" {
		g.queuedDirection = ""
		return nil
	}
	if network.cooldown > 0 {
		return nil
	}
	if g.queuedDirection == "" {
		g.queuedDirection = movementDirection(ebiten.IsKeyPressed)
	}
	if g.queuedDirection != "" {
		action := "move"
		if network.view.Role == domain.Navigator {
			action = "hint"
		}
		g.sendRoomCommand(action, g.queuedDirection)
		g.queuedDirection = ""
	}
	return nil
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (g *Game) sendRoomCommand(action, direction string) {
	network := g.multiplayer
	if action == "restart" && network.view.Role != domain.Explorer {
		network.errorMessage, network.errorTicks = "The explorer starts the next maze.", 180
		return
	}
	if network.client.Send(protocol.ClientMessage{Type: action, Direction: direction}) {
		network.cooldown = 13
	}
}

func multiplayerButton(index int) image.Rectangle {
	top := mazeHeight*tileSize + 94
	if index == 4 {
		return image.Rect(184, top, 232, top+18)
	}
	return image.Rect(8+44*index, top, 48+44*index, top+18)
}

func (g *Game) multiplayerClick() string {
	var points []image.Point
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		points = append(points, g.pointerPosition(x, y))
	}
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		points = append(points, g.pointerPosition(x, y))
	}
	actions := []string{"up", "left", "down", "right", "restart"}
	for _, point := range points {
		for index, action := range actions {
			if point.In(multiplayerButton(index)) {
				return action
			}
		}
	}
	return ""
}

func (g *Game) drawMultiplayer(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 24, G: 26, B: 42, A: 255})
	network := g.multiplayer
	if view := network.view; view != nil {
		found := g.friendFound()
		for y, row := range view.Cells {
			for x, cell := range row {
				if cell == nil {
					continue
				}
				tile := g.hedgeTile
				if !g.useHedge {
					tile = g.woodTile
				}
				if *cell {
					tile = g.dirtTile
				}
				options := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
				if view.Role == domain.Explorer && !found {
					if view.LightRemainingMS > 0 {
						options.ColorScale.Scale(1.12, 1.05, 0.92, 1)
					} else {
						options.ColorScale.Scale(0.45, 0.47, 0.60, 1)
					}
				}
				options.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
				screen.DrawImage(tile, options)
			}
		}
		for _, lantern := range view.Lanterns {
			drawSprite(screen, g.lanternSprite, float64(lantern.X*tileSize), float64(lantern.Y*tileSize))
		}
		if friend := view.Friend; friend != nil {
			x := float64(friend.X * tileSize)
			if found {
				x += 5
			}
			drawSprite(screen, g.friendSprite, x, float64(friend.Y*tileSize))
		}
		x, y := g.walk.position()
		playerX := math.Round(x * tileSize)
		if found {
			playerX -= 5
		}
		drawSprite(screen, g.playerSprites[g.walk.frame()], playerX, math.Round(y*tileSize))
		g.drawWallBubble(screen, x, y)
	}
	g.drawMultiplayerHUD(screen)
}

func (g *Game) drawMultiplayerHUD(screen *ebiten.Image) {
	network := g.multiplayer
	top := mazeHeight * tileSize
	background := color.RGBA{R: 35, G: 31, B: 45, A: 255}
	ink := color.RGBA{R: 51, G: 39, B: 24, A: 255}
	gold := color.RGBA{R: 239, G: 193, B: 109, A: 255}
	textColor := color.RGBA{R: 255, G: 241, B: 211, A: 255}
	vector.FillRect(screen, 0, float32(top), mazeWidth*tileSize, multiplayerHUDHeight, background, false)
	vector.FillRect(screen, 0, float32(top), mazeWidth*tileSize, 1, gold, false)
	phase := network.connection
	if phase == "" {
		phase = "Connecting to room..."
	}
	if view := network.view; view != nil {
		drawHUDText(screen, fmt.Sprintf("%s | %s", view.Role, view.Code), 8, top+7, gold)
		if network.connected {
			switch view.Phase {
			case "waiting":
				phase = "Waiting for navigator..."
			case "paused":
				phase = "Paused: partner disconnected."
			case "playing":
				phase = "Find your lost friend together."
			case "won":
				phase = "Friend found! Both safe."
			}
		}
		drawHUDText(screen, fmt.Sprintf("Moves: %d | Lanterns: %d/%d", view.Moves, view.LanternsCollected, view.LanternTotal), 8, top+44, textColor)
		light := "off"
		if view.LightRemainingMS > 0 {
			light = fmt.Sprintf("%ds", (view.LightRemainingMS+999)/1000)
		}
		drawHUDText(screen, fmt.Sprintf("Light: %s | Sight: %d", light, view.VisionRadius), 8, top+55, textColor)
		hint := "Navigator: waiting for directions"
		if view.Hint != "" {
			hint = "Navigator: " + view.Hint
		}
		drawHUDText(screen, hint, 8, top+66, gold)
		controls := "Arrows/WASD: move | T: walls"
		if view.Role == domain.Navigator {
			controls = "Arrows/WASD: send directions"
		}
		drawHUDText(screen, controls, 8, top+79, textColor)
	}
	if network.errorTicks > 0 {
		phase = network.errorMessage
	}
	drawHUDText(screen, wrapMessage(phase, 36), 8, top+20, textColor)
	for index, label := range []string{"UP", "LEFT", "DOWN", "RIGHT", "New[R]"} {
		box := multiplayerButton(index)
		vector.FillRect(screen, float32(box.Min.X), float32(box.Min.Y), float32(box.Dx()), float32(box.Dy()), gold, false)
		drawHUDText(screen, label, box.Min.X+(box.Dx()-len(label)*6)/2, box.Min.Y+5, ink)
	}
}
