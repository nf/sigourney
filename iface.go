package main

import "math"

const (
	waveHz             = 44100
	waveAmp            = 32768
	sampleSize float64 = 2 * math.Pi / waveHz
)

type Sample float64

type Processor interface {
	Process([]Sample)
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

type Amp struct {
	sig, ctl Processor
	buf      []Sample // for ctl
}

func (a *Amp) Process(s []Sample) {
	if a.buf == nil {
		a.buf = make([]Sample, len(s))
	}
	a.ctl.Process(a.buf)
	a.sig.Process(s)
	for i := range s {
		s[i] *= a.buf[i]
	}
}

type Env struct {
	attack, decay int
	p             int
}

func (e *Env) Process(s []Sample) {
	for i := range s {
		switch {
		case e.p <= e.attack:
			s[i] = Sample(e.p) / Sample(e.attack)
		default:
			s[i] = 1.0 - Sample(e.p-e.attack)/Sample(e.decay)
		}
		e.p++
		if e.p > e.attack+e.decay {
			e.p = 0
		}
	}
}
