package main

import (
	"bytes"
	"encoding/binary"
	"github.com/nf/wav"
	"os"
)

var song = []string{
	"C-3", "C#3", "D-3", "D#3", "E-3", "F-3",
	"F#3", "G-3", "G#3", "A-3", "A#3", "B-3",
	"C-4", "C#4", "D-4", "D#4", "E-4", "F-4",
	"F#4", "G-4", "G#4", "A-4", "A#4", "B-4",
	"C-5", "C#5", "D-5", "D#5", "E-5", "F-5",
	"F#5", "G-5", "G#5", "A-5", "A#5", "B-5",
	"C-6",
}

const (
	noteDuration = waveHz / 0.1
	songLength   = waveHz * 10
)

func main() {
	// build song
	var q []Note
	for _, n := range song {
		q = append(q, Note{NoteToFreq(n), 1.0, 8000}, Note{0, 0, 2000})
	}

	osc := NewModule(&Oscillator{}, &NoteLane{Q: q})
	env := NewModule(&AmpEnvelope{200, 1000}, &NoteLane{Q: q})

	in := make(chan Buf)
	out := env.Run(osc.Run(in))

	in <- make(Buf, 1000)

	// play song, writing to buf
	var buf bytes.Buffer
	played := 0
	for b := range out {
		for _, v := range b {
			binary.Write(&buf, binary.LittleEndian, int16(v*waveAmp))
		}
		played += len(b)
		if played < songLength {
			in <- b
		} else {
			break
		}
	}

	f, err := os.Open("test.wav", os.O_WRONLY|os.O_CREAT|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := &wav.File{
		SampleRate:      waveHz,
		SignificantBits: 16,
		Channels:        1,
	}
	if err := w.WriteData(f, buf.Bytes()); err != nil {
		panic(err)
	}
}
