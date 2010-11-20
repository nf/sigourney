package main

import (
	"math"
)

type Processor interface {
	Process(Sample, Instructor) Sample
}

type Instructor interface {
	Command() interface{}
	Advance(int) int
}

type Module struct {
	p Processor
	i Instructor
}

func NewModule(p Processor, i Instructor) *Module {
	return &Module{p, i}
}

func (m *Module) Run(in chan Buf) chan Buf {
	out := make(chan Buf)
	go func() {
		for b := range in {
			for i := range b {
				b[i] = m.p.Process(b[i], m.i)
			}
			out <- b
		}
	}()
	return out
}

type Oscillator struct{}

func (osc *Oscillator) Process(s Sample, i Instructor) Sample {
	n := i.Command().(Note)
	p := i.Advance(1)
	s = Sample(math.Sin(n.Freq*sampleSize*float64(p)) * n.Amp)
	return s
}

type AmpEnvelope struct {
	Attack  int
	Release int
}

func (amp *AmpEnvelope) Process(s Sample, i Instructor) Sample {
	n := i.Command().(Note)
	p := i.Advance(1)
	var f float64 = 1.0
	if amp.Attack > p {
		a := float64(amp.Attack)
		f = 1.0 - (a-float64(p))/a
	}
	if amp.Release > n.Duration-p {
		f = float64(n.Duration-p) / float64(amp.Release)
	}
	s = Sample(float64(s) * f)
	return s
}

type NoteLane struct {
	Q  []Note
	qp int // queue position
	n  Note
	np int // note position
}

func (l *NoteLane) Command() interface{} {
	if l.np >= l.n.Duration {
		for {
			if l.qp >= len(l.Q) {
				// end of notes
				return Note{}
			}
			l.n = l.Q[l.qp]
			if l.np < l.n.Duration {
				break // play note
			}
			// finished note
			l.qp++
			l.np = 0
		}
	}
	return l.n
}

func (l *NoteLane) Advance(i int) int {
	j := l.np
	l.np += i
	return j
}
