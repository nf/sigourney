package main

import (
	"log"
	"os"

	"code.google.com/p/portaudio-go/portaudio"
)

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	e := NewEngine()

	oscMod := &SimpleOsc{}
	oscMod.SetInput("pitch", Value(-0.1))

	oscModAmp := &Amp{}
	oscModAmp.SetInput("car", oscMod)
	oscModAmp.SetInput("mod", Value(0.1))

	osc := &SimpleOsc{}
	osc.SetInput("pitch", oscModAmp)

	envMod := &SimpleOsc{}
	envMod.SetInput("pitch", Value(-1))

	envModAmp := &Amp{}
	envModAmp.SetInput("car", envMod)
	envModAmp.SetInput("mod", Value(0.02))

	envModSum := &Sum{}
	envModSum.SetInput("car", envModAmp)
	envModSum.SetInput("mod", Value(0.021))

	env := &Env{}
	env.SetInput("att", Value(0.0001))
	env.SetInput("dec", envModSum)

	amp := &Amp{}
	amp.SetInput("car", osc)
	amp.SetInput("mod", env)

	ampAmp := &Amp{}
	ampAmp.SetInput("car", amp)
	ampAmp.SetInput("mod", Value(0.5))

	e.SetInput("", ampAmp)

	if err := e.Start(); err != nil {
		log.Println(err)
		return
	}

	os.Stdout.Write([]byte("Press enter to stop...\n"))
	os.Stdin.Read([]byte{0})

	if err := e.Stop(); err != nil {
		log.Println(err)
	}
}

func NewEngine() *Engine {
	return &Engine{buf: make([]Sample, nSamples), done: make(chan error)}
}

type Engine struct {
	buf  []Sample
	root Processor
	done chan error
}

func (e *Engine) processAudio(_, out []int16) {
	e.root.Process(e.buf)
	for i := range e.buf {
		out[i] = int16(e.buf[i] * waveAmp)
	}
}

func (e *Engine) SetInput(_ string, p Processor) {
	e.root = p
}

func (e *Engine) Start() error {
	stream, err := portaudio.OpenDefaultStream(0, 1, waveHz, nSamples, e.processAudio)
	if err != nil {
		return err
	}
	errc := make(chan error)
	go func() {
		err = stream.Start()
		errc <- err
		if err != nil {
			return
		}
		<-e.done
		err = stream.Stop()
		if err == nil {
			err = stream.Close()
		}
		e.done <- err
	}()
	return <-errc
}

func (e *Engine) Stop() error {
	e.done <- nil
	return <-e.done
}
