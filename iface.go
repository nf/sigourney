package main

const (
	nChannels = 1
	nSamples  = 256 * nChannels
)

const (
	waveHz  = 44100
	waveAmp = 32768
)

type Sample float64

type Processor interface {
	Process([]Sample)
}

type Sink interface {
	SetInput(name string, g Processor)
}
