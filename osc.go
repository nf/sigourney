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

type Note struct {
	Freq float64
	Amp float64
	Duration int
}
type Sample float64
type Buf []Sample

type Oscillator struct {
	q []Note
	p int // note play position
	freq, amp float64
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
				s := math.Sin(float64(osc.p)*n.Freq*sampleSize)
				s *= n.Amp
				m[i] = Sample(s)
				osc.p++
			}
			out <- m
		}
	}
}

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

func main() {
	osc := new(Oscillator)

	in := make(chan interface{})
	out := osc.Run(in)

	// song
	in <- Note{440, 0.8, 11025}
	in <- Note{880, 0.6, 11025}
	in <- Note{220, 0.4, 11025}
	in <- Note{440, 0.2, 11025}

	in <- make(Buf, 44100)

	// fill buf with wave data
	var buf bytes.Buffer
	for b := range out {
		for _, v := range b {
			binary.Write(&buf, binary.LittleEndian, int16(v * waveAmp))
		}
		close(out) // FIXME
	}

	f, err := os.Open("test.wav", os.O_WRONLY|os.O_CREAT|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := &wav.File{
		SampleRate: waveHz,
		SignificantBits: 16,
		Channels: 1,
	}
	if err := w.WriteData(f, buf.Bytes()); err != nil {
		panic(err)
	}
}
