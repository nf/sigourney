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

type Processor interface {
	Process([]Sample)
	Tick()
}

type SimpleOsc struct {
	freq float64 // Hz
	pos  int
}

func (o *SimpleOsc) Process(s []Sample) {
	fs := o.freq * sampleSize
	p := float64(o.pos)
	for i := range s {
		s[i] = Sample(math.Sin(fs * p))
		p++
	}
}

func (o *SimpleOsc) Tick() {
	o.pos += nSamples
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

func (*Amp) Tick() {}

type Env struct {
	attack, decay int
	pos           int
}

func (e *Env) Process(s []Sample) {
	p := Sample(e.pos)
	att, dec := Sample(e.attack), Sample(e.decay)
	period := Sample(e.attack + e.decay)
	for i := range s {
		if p <= att {
			s[i] = p / att
		} else {
			s[i] = 1.0 - (p-att)/dec
		}
		p++
		if p > period {
			p = 0
		}
	}
}

func (e *Env) Tick() {
	e.pos += nSamples
	e.pos %= e.attack + e.decay
}
