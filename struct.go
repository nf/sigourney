package main

import "math"

const (
	waveHz  = 44100
	waveAmp = 32768
)

var (
	sampleSize float64 = 2 * math.Pi / waveHz
)

type Note struct {
	Freq     float64
	Amp      float64
	Duration int
}
type Sample float64
type Buf []Sample

