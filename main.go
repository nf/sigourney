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

	osc2 := &SimpleOsc{}
	osc2.SetInput("pitch", Value(-0.1))

	amp4 := &Amp{}
	amp4.SetInput("car", osc2)
	amp4.SetInput("mod", Value(0.1))

	osc := &SimpleOsc{}
	osc.SetInput("pitch", amp4)

	env := &Env{}
	env.SetInput("att", Value(1))
	env.SetInput("dec", Value(1))

	amp3 := &Amp{}
	amp3.SetInput("car", env)
	amp3.SetInput("mod", Value(0.1))

	env2 := &Env{}
	env2.SetInput("att", Value(0.001))
	env2.SetInput("dec", amp3)

	amp := &Amp{}
	amp.SetInput("car", osc)
	amp.SetInput("mod", env2)

	amp2 := &Amp{}
	amp2.SetInput("car", amp)
	amp2.SetInput("mod", Value(0.5))

	e.SetInput("", amp2)

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
