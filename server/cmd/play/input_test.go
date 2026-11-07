package main

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

func TestReadCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"WASD up", "w", "up"},
		{"WASD down", "s", "down"},
		{"WASD left", "a", "left"},
		{"WASD right", "d", "right"},
		{"uppercase up", "W", "up"},
		{"uppercase down", "S", "down"},
		{"uppercase left", "A", "left"},
		{"uppercase right", "D", "right"},
		{"arrow up", "\x1b[A", "up"},
		{"arrow down", "\x1b[B", "down"},
		{"arrow right", "\x1b[C", "right"},
		{"arrow left", "\x1b[D", "left"},
		{"application arrow up", "\x1bOA", "up"},
		{"application arrow down", "\x1bOB", "down"},
		{"application arrow right", "\x1bOC", "right"},
		{"application arrow left", "\x1bOD", "left"},
		{"modified arrow", "\x1b[1;5C", "right"},
		{"quit", "q", "quit"},
		{"uppercase quit", "Q", "quit"},
		{"Ctrl+C", "\x03", "quit"},
		{"Ctrl+D", "\x04", "quit"},
		{"Enter", "\r", ""},
		{"newline", "\n", ""},
		{"unknown key", "x", ""},
		{"unsupported escape", "\x1b[3~", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := readCommand(bufio.NewReader(strings.NewReader(test.input)))
			if err != nil {
				t.Fatalf("failed to read command: %v", err)
			}
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestReadCommandPreservesFollowingKeys(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\x1b[Ada\n\x1bq"))
	for _, expected := range []string{"up", "right", "left", "", "", "quit"} {
		actual, err := readCommand(reader)
		if err != nil || actual != expected {
			t.Fatalf("expected %q, got %q with error %v", expected, actual, err)
		}
	}
	if _, err := readCommand(reader); err != io.EOF {
		t.Errorf("expected EOF after all commands, got %v", err)
	}
}

func TestReadCommandHandlesIncompleteInput(t *testing.T) {
	for _, input := range []string{"", "\x1b", "\x1b[", "\x1bO", "\x1b[1;"} {
		if _, err := readCommand(bufio.NewReader(strings.NewReader(input))); err != io.EOF {
			t.Errorf("input %q: expected EOF, got %v", input, err)
		}
	}
}
