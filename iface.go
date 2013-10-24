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
	f   Sample // 0.1/oct, 0 == 440Hz
	pos int
}

func (o *SimpleOsc) Process(s []Sample) {
	step := 440 * math.Exp2(float64(o.f)*10) * sampleSize
	p := float64(o.pos) * step
	for i := range s {
		s[i] = Sample(math.Sin(p))
		p += step
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
