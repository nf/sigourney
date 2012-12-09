package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"log"
	"time"
)

func main() {
	a := &audio{
		p: &SimpleOsc{freq: 220},
	}
	stream, err := portaudio.OpenDefaultStream(0, 1, waveHz, nSamples, a)
	if err != nil {
		log.Fatal(err)
	}
	err = stream.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = stream.Stop()
	if err != nil {
		log.Fatal(err)
	}
	err = stream.Close()
	if err != nil {
		log.Fatal(err)
	}
}

type audio struct{
	p Processor
}

const (
	nChannels = 1
	nSamples = 256 * nChannels
)

var buf = make([]Sample, nSamples)

func (a *audio) ProcessAudio(_, out []int16) {
	a.p.Process(buf)
	for i := range buf {
		out[i] = int16(buf[i]*waveAmp)
	}
}
