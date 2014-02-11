package main

import "math"

const (
	nChannels = 1
	nSamples  = 256 * nChannels
)

const (
	waveHz             = 44100
	waveAmp            = 32768
	sampleSize float64 = 2 * math.Pi / waveHz
)

type Sample float64

type Ticker interface {
	Tick()
}

type Processor interface {
	Process([]Sample)
}

type Sink interface {
	SetInput(name string, g Processor)
}
