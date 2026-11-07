package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/putriindaht/leftnt/server/internal/domain"
)

func (g *Game) restart() error {
	state, err := domain.NewSoloGame(mazeWidth, mazeHeight, domain.DefaultSoloConfig())
	if err != nil {
		return err
	}
	g.resetState(state)
	ebiten.SetWindowTitle("LEFTN'T - Ebitengine")
	return nil
}

// Keep cached images and the chosen wall style; reset the previous run's state.
func (g *Game) resetState(state *domain.SoloGame) {
	g.state = state
	g.walk = newWalkAnimation(state.Player.Position)
	g.queuedDirection = ""
	g.blockedDirection = ""
	g.winAnnounced = false
	g.feedback = wallFeedback{}
}
