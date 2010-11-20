package main

import (
	"math"
)

type Oscillator struct {
	q []Note
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
				n := osc.note()
				s := math.Sin(float64(osc.p) * n.Freq * sampleSize)
				s *= n.Amp
				m[i] = Sample(s)
				osc.p++
			}
			out <- m
		}
	}
}

// note returns the currently playing note
func (osc *Oscillator) note() Note {
	var n Note
	for {
		if len(osc.q) == 0 {
			n = Note{}
			break
		}
		n = osc.q[0]
		if osc.p < n.Duration {
			// still playing this note
			break
		}
		// finished current note, shift off
		osc.q = osc.q[1:]
		osc.p = 0
	}
	return n
}
