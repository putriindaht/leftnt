package main

import (
	"math/rand/v2"
	"strings"
)

const (
	wallMessageDuration = 180
	wallMessageCooldown = 45
)

var wallMessages = []string{
	"That's a wall, bestie.",
	"Your other left. The OTHER one.",
	"Bold strategy. The wall disagrees.",
	"The wall won that conversation.",
	"Found a wall. Your friend is elsewhere.",
	"Shortcut denied. Please use a path.",
	"Excellent wall inspection. Very thorough.",
	"GPS has left the garden.",
	"Even the lantern saw that coming.",
	"This is a maze, not a demolition job.",
	"The wall remains undefeated.",
	"Plot twist: still a wall.",
	"No secret door. Just commitment.",
	"The path is the part without walls.",
}

type wallFeedback struct {
	message   string
	remaining int
	cooldown  int
}

func (f *wallFeedback) advance() {
	if f.remaining > 0 {
		f.remaining--
	}
	if f.cooldown > 0 {
		f.cooldown--
	}
}

func (f *wallFeedback) bump() {
	if f.cooldown > 0 {
		return
	}
	f.message = nextWallMessage(f.message)
	f.remaining = wallMessageDuration
	f.cooldown = wallMessageCooldown
}

func nextWallMessage(previous string) string {
	candidates := make([]string, 0, len(wallMessages))
	for _, message := range wallMessages {
		if message != previous {
			candidates = append(candidates, message)
		}
	}
	return candidates[rand.IntN(len(candidates))]
}

func wrapMessage(message string, columns int) string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(message) {
		if line != "" && len(line)+1+len(word) > columns {
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
