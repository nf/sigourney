package main

import (
	"bytes"
	"encoding/binary"
	"github.com/nf/wav"
	"os"
)

func main() {
	osc := new(Oscillator)

	in := make(chan interface{})
	out := osc.Run(in)

	// song
	for _, n := range []string{"C-4", "C#4", "D-4", "D-5"} {
		in <- Note{NoteToFreq(n), 1.0, 11025}
	}

	in <- make(Buf, 44100)

	// fill buf with wave data
	var buf bytes.Buffer
	for b := range out {
		for _, v := range b {
			binary.Write(&buf, binary.LittleEndian, int16(v*waveAmp))
		}
		close(out) // FIXME
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
