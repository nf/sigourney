package main

import (
	"log"
	"os"

	"code.google.com/p/portaudio-go/portaudio"
)

func main() {
	// Collect processors so we can call tick on each once per frame.
	var ps []Processor
	track := func(p Processor) Processor {
		ps = append(ps, p)
		return p
	}

	// Build signal chain.
	var p Processor
	p = track(&SimpleOsc{})
	p = track(&Amp{
		sig: p,
		ctl: track(&Env{attack: waveHz / 100, decay: waveHz / 3}),
		//ctl: track(&Value{1}),
	})

	a := &audio{ps: ps, root: p, buf: make([]Sample, nSamples)}
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
	ps   []Processor
	root Processor
	buf  []Sample
}

func (a *audio) ProcessAudio(_, out []int16) {
	a.root.Process(a.buf)
	for i := range a.buf {
		out[i] = int16(a.buf[i] * waveAmp)
	}
	for _, p := range a.ps {
		p.Tick()
	}
}
