package main

import (
	"fmt"
	"math"
)

const (
	waveHz  = 44100
	waveAmp = 32768
)

var (
	sampleSize float64 = 2 * math.Pi / waveHz
)

type Freq float64
type Amp float64
type Sample int16
type Buf []Sample

func osc(in chan interface{}) chan Buf {
	out := make(chan Buf)
	go func() {
		var p int
		var freq float64
		var amp float64
		for {
			switch m := (<-in).(type) {
			case Freq:
				freq = float64(m) * sampleSize
			case Amp:
				amp = float64(m) * waveAmp
			case Buf:
				for i := range m {
					f := math.Sin(float64(p)*freq) * amp
					m[i] = Sample(f)
					p++
				}
				out <- m
			}
		}
	}()
	return out
}

func main() {
	in := make(chan interface{})
	out := osc(in)
	in <- Amp(0.8)
	in <- Freq(440)
	in <- make(Buf, 100)
	b := <-out
	for i, v := range b {
		fmt.Println(i, v)
	}
}
