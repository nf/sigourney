package main

import (
	"bytes"
	"encoding/binary"
	"github.com/nf/wav"
	"math"
	"os"
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
	w := &wav.File{
		SampleRate: waveHz,
		SignificantBits: 16,
		Channels: 1,
	}

	in := make(chan interface{})
	out := osc(in)
	in <- Amp(0.8)
	in <- Freq(440)
	in <- make(Buf, 44100)

	// fill buf with wave data
	var buf bytes.Buffer
	for b := range out {
		for _, v := range b {
			binary.Write(&buf, binary.LittleEndian, v)
		}
		close(out) // FIXME
	}

	f, err := os.Open("test.wav", os.O_WRONLY|os.O_CREAT|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = w.WriteData(f, buf.Bytes())
	if err != nil {
		panic(err)
	}
}
