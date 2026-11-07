package main

import "bufio"

// readCommand translates keyboard bytes into the directions used by the domain.
func readCommand(reader *bufio.Reader) (string, error) {
	key, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	switch key {
	case 'w', 'W':
		return "up", nil
	case 's', 'S':
		return "down", nil
	case 'a', 'A':
		return "left", nil
	case 'd', 'D':
		return "right", nil
	case 'q', 'Q', 3, 4:
		// Raw mode delivers Ctrl+C and Ctrl+D as bytes.
		return "quit", nil
	case '\x1b':
		return readArrowCommand(reader)
	default:
		return "", nil
	}
}

func readArrowCommand(reader *bufio.Reader) (string, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return "", err
	}
	// Terminals send arrows as ESC [ A/B/C/D or ESC O A/B/C/D.
	if prefix != '[' && prefix != 'O' {
		return "", reader.UnreadByte()
	}

	for {
		key, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		if key == 3 || key == 4 {
			return "quit", nil
		}
		// Skip optional parameters, such as those sent with Ctrl+arrow.
		if key < 0x40 || key > 0x7e {
			continue
		}
		switch key {
		case 'A':
			return "up", nil
		case 'B':
			return "down", nil
		case 'C':
			return "right", nil
		case 'D':
			return "left", nil
		default:
			return "", nil
		}
	}
}
