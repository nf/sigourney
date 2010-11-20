package main

import (
	"math"
)

type Oscillator struct {
	q []Note
	n Note // current note
	p int // note play position
}

func (osc *Oscillator) Run(in chan interface{}) chan Buf {
	out := make(chan Buf)
	go osc.play(in, out)
	return out
}

func (osc *Oscillator) play(in chan interface{}, out chan Buf) {
	for {
		switch m := (<-in).(type) {
		case Note:
			osc.q = append(osc.q, m)
		case Buf:
			for i := range m {
				m[i] = osc.sample()
			}
			out <- m
		}
	}
}

// Sample returns the current sample and advances the playhead
func (osc *Oscillator) sample() Sample {
	if osc.p >= osc.n.Duration {
		for {
			if len(osc.q) == 0 {
				return 0
			}
			osc.n = osc.q[0]
			if osc.p < osc.n.Duration {
				break // play this note
			}
			// finished current note, shift off
			osc.q = osc.q[1:]
			osc.p = 0
		}
	}
	osc.p++
	s := math.Sin(float64(osc.p) * osc.n.Freq * sampleSize) * osc.n.Amp
	return Sample(s)
}
