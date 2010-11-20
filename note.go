package main

import (
	"math"
	"strconv"
)

// NoteToFreq parses a note of the form C-4, D#5, etc and 
// returns its frequency in Hz
func NoteToFreq(note string) float64 {
	if len(note) != 3 {
		return 0
	}

	var cents int
	switch note[0] {
	case 'A':
		cents = -300
	case 'B':
		cents = -100
	case 'C':
	case 'D':
		cents = 200
	case 'E':
		cents = 400
	case 'F':
		cents = 500
	case 'G':
		cents = 600
	default:
		return 0
	}

	switch note[1] {
	case '-':
	case '#':
		cents += 100
	default:
		return 0
	}

	if oct, err := strconv.Atoi(note[2:]); err == nil {
		cents += (oct - 4) * 1200
	} else {
		return 0
	}

	return 440 * math.Exp2(float64(cents)/1200)
}
