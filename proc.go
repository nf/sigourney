package main

import "math"

func NewOsc() *Osc {
	o := &Osc{}
	newSink(&o.sink, "pitch", &o.pitch)
	return o
}

type Osc struct {
	sink
	pitch source // 0.1/oct, 0 == 440Hz

	pos float64
}

func (o *Osc) Process(s []Sample) {
	pitch := o.pitch.Process()
	p := o.pos
	for i := range s {
		s[i] = Sample(math.Sin(p * 2 * math.Pi))
		hz := 440 * math.Exp2(float64(pitch[i])*10)
		p += hz / waveHz
		if p > 100 {
			p -= 100
		}
	}
	o.pos = p
}

func NewAmp() *Amp {
	a := &Amp{}
	newSink(&a.sink, "car", &a.car, "mod", &a.mod)
	return a
}

type Amp struct {
	sink
	car Processor
	mod source
}

func (a *Amp) Process(s []Sample) {
	a.car.Process(s)
	m := a.mod.Process()
	for i := range s {
		s[i] *= m[i]
	}
}

func NewSum() *Sum {
	s := &Sum{}
	newSink(&s.sink, "car", &s.car, "mod", &s.mod)
	return s
}

type Sum struct {
	sink
	car Processor
	mod source
}

func (a *Sum) Process(s []Sample) {
	a.car.Process(s)
	m := a.mod.Process()
	for i := range s {
		s[i] += m[i]
	}
}

func NewEnv() *Env {
	e := &Env{}
	newSink(&e.sink, "att", &e.att, "dec", &e.dec)
	return e
}

type Env struct {
	sink
	att, dec source

	down bool
	v    Sample
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

type Value Sample

func (v Value) Process(s []Sample) {
	for i := range s {
		s[i] = Sample(v)
	}
}

type sink struct {
	inputs map[string]interface{}
}

func newSink(s *sink, args ...interface{}) {
	s.inputs = make(map[string]interface{})
	if len(args)%2 != 0 {
		panic("odd number of args")
	}
	for i := 0; i < len(args); i++ {
		name, ok := args[i].(string)
		if !ok {
			panic("invalid args; expected string")
		}
		i++
		s.inputs[name] = args[i]
	}
}

func (s *sink) Input(name string, p Processor) {
	if s.inputs == nil {
		panic("no inputs registered")
	}
	i, ok := s.inputs[name]
	if !ok {
		panic("bad input name: " + name)
	}
	switch v := i.(type) {
	case *Processor:
		*v = p
	case *source:
		if (*v).b == nil {
			(*v).b = make([]Sample, nSamples)
		}
		(*v).p = p
	default:
		panic("bad input type")
	}
}

type source struct {
	p Processor
	b []Sample
}

func (s *source) Process() []Sample {
	s.p.Process(s.b)
	return s.b
}
