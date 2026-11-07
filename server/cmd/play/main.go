package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/putriindaht/leftnt/server/internal/domain"
	"golang.org/x/term"
)

func main() {
	maze, err := domain.GenerateMaze(15, 11)
	if err != nil {
		fmt.Println("Failed to generate maze:", err)
		return
	}

	player := domain.Player{
		Position: domain.Position{X: 1, Y: 1},
	}

	goal, err := maze.RandomGoal(player.Position)
	if err != nil {
		fmt.Println("Failed to select goal:", err)
		return
	}

	var output io.Writer = os.Stdout
	inputFD := int(os.Stdin.Fd())
	if term.IsTerminal(inputFD) {
		state, err := term.MakeRaw(inputFD)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to configure terminal:", err)
			return
		}
		defer term.Restore(inputFD, state)

		if term.IsTerminal(int(os.Stdout.Fd())) {
			// Preserve normal line breaks while the terminal is in raw mode.
			output = term.NewTerminal(os.Stdout, "")
		}
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Fprintln(output, "LEFTN'T — Terminal Practice")
	fmt.Fprintln(output, "P = player | G = goal | . = path | # = wall")
	fmt.Fprintln(output, "Use arrow keys or W/A/S/D to move. No Enter needed.")
	fmt.Fprintln(output, "Press Q or Ctrl+C to quit.")

	drawMaze(output, maze, player, goal)

	for {
		direction, err := readCommand(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(output, "Failed to read input:", err)
			}
			return
		}

		switch direction {
		case "quit":
			return
		case "":
			continue
		}

		if !player.Move(direction, maze) {
			fmt.Fprintln(output, "You can't move there!")
		}

		drawMaze(output, maze, player, goal)

		if player.Position == goal {
			fmt.Fprintln(output, "Goal reached! Your sense of direction survived.")
			return
		}
	}
}

func drawMaze(
	output io.Writer,
	maze domain.Maze,
	player domain.Player,
	goal domain.Position,
) {
	fmt.Fprintln(output)

	for y, row := range maze.Cells {
		for x, walkable := range row {
			position := domain.Position{X: x, Y: y}

			switch {
			case player.Position == position:
				fmt.Fprint(output, "P ")
			case goal == position:
				fmt.Fprint(output, "G ")
			case walkable:
				fmt.Fprint(output, ". ")
			default:
				fmt.Fprint(output, "# ")
			}
		}

		fmt.Fprintln(output)
	}

	fmt.Fprintf(output, "\nPosition: (%d, %d)\n", player.Position.X, player.Position.Y)
}
