package main

import (
	"math"
)

const (
	waveHz             = 44100
	waveAmp            = 32768
	sampleSize float64 = 2 * math.Pi / waveHz
)

type Sample float64

type TriggerType int

const (
	Panic TriggerType = iota
	NoteOn
	NoteOff
)

type Processor interface {
	Process([]Sample)
	Trigger(TriggerType, interface{})
}

type SimpleOsc struct {
	freq float64 // Hz
	pos  int
}

func (o *SimpleOsc) Process(s []Sample) {
	for i := range s {
		s[i] = Sample(math.Sin(o.freq * sampleSize * float64(o.pos)))
		o.pos++
	}
}

func (o *SimpleOsc) Trigger(t TriggerType, note interface{}) {
	o.freq = 440
	o.pos = 0
}
