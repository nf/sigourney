package main

import "math"

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
	att, dec *Source
	down     bool
	v        Sample
}

func (e *Env) Process(s []Sample) {
	att, dec := e.att.Process(), e.dec.Process()
	v := e.v
	for i := range s {
		if e.down {
			if d := dec[i]; d > 0 {
				v -= 1 / (d * waveHz * 10)
			}
		} else {
			if a := att[i]; a > 0 {
				v += 1 / (a * waveHz * 10)
			}
		}
		if v <= 0 {
			v = 0
			e.down = false
		} else if v >= 1 {
			v = 1
			e.down = true
		}
		s[i] = v
	}
	e.v = v
}

func (e *Env) SetInput(name string, p Processor) {
	switch name {
	case "att":
		if e.att == nil {
			e.att = NewSource(p)
		} else {
			e.att.SetInput("", p)
		}
	case "dec":
		if e.dec == nil {
			e.dec = NewSource(p)
		} else {
			e.dec.SetInput("", p)
		}
	default:
		panic("bad input")
	}
}

type Value Sample

func (v Value) Process(s []Sample) {
	for i := range s {
		s[i] = Sample(v)
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
