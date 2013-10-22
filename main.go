package main

import (
	"log"
	"os"

	"code.google.com/p/portaudio-go/portaudio"
)

const (
	nChannels = 1
	nSamples  = 256 * nChannels
)

func main() {
	var p Processor
	p = &SimpleOsc{freq: 220}
	p = &Amp{sig: p, ctl: &Env{attack: waveHz / 100, decay: waveHz / 3}}

	a := &audio{p: p, buf: make([]Sample, nSamples)}
	stream, err := portaudio.OpenDefaultStream(0, 1, waveHz, nSamples, a)
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan bool)
	go func() {
		err = stream.Start()
		if err != nil {
			log.Fatal(err)
		}
		err = stream.Start()
		<-done
		err = stream.Stop()
		if err != nil {
			log.Fatal(err)
		}
		err = stream.Close()
		if err != nil {
			log.Fatal(err)
		}
		done <- true
	}()
	os.Stdout.Write([]byte("Press enter to stop...\n"))
	os.Stdin.Read([]byte{0})
	done <- true
	<-done
}

type audio struct {
	p   Processor
	buf []Sample
	i   int
}

func (a *audio) ProcessAudio(_, out []int16) {
	a.p.Process(a.buf)
	for i := range a.buf {
		out[i] = int16(a.buf[i] * waveAmp)
	}
}
