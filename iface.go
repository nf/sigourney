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
	car Processor
	mod *Source
}

func (a *Amp) Process(s []Sample) {
	a.car.Process(s)
	m := a.mod.Process()
	for i := range s {
		s[i] *= m[i]
	}
}

func (a *Amp) SetInput(name string, p Processor) {
	switch name {
	case "car":
		a.car = p
	case "mod":
		if a.mod == nil {
			a.mod = NewSource(p)
		} else {
			a.mod.SetInput("", p)
		}
	default:
		panic("bad input")
	}
}

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

type Value struct {
	v Sample
}

func (v *Value) Process(s []Sample) {
	for i := range s {
		s[i] = v.v
	}
}

func NewSource(p Processor) *Source {
	return &Source{p: p}
}

type Source struct {
	p Processor
	b []Sample
}

func (s *Source) Process() []Sample {
	if s.b == nil {
		s.b = make([]Sample, nSamples)
	}
	s.p.Process(s.b)
	return s.b
}

func (s *Source) SetInput(_ string, p Processor) {
	s.p = p
}
