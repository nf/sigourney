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

	oscMod := NewOsc()
	oscMod.Input("pitch", Value(-0.1))

	oscModAmp := NewAmp()
	oscModAmp.Input("car", oscMod)
	oscModAmp.Input("mod", Value(0.1))

	osc := NewOsc()
	osc.Input("pitch", oscModAmp)

	envMod := NewOsc()
	envMod.Input("pitch", Value(-1))

	envModAmp := NewAmp()
	envModAmp.Input("car", envMod)
	envModAmp.Input("mod", Value(0.02))

	envModSum := NewSum()
	envModSum.Input("car", envModAmp)
	envModSum.Input("mod", Value(0.021))

	env := NewEnv()
	env.Input("att", Value(0.0001))
	env.Input("dec", envModSum)

	amp := NewAmp()
	amp.Input("car", osc)
	amp.Input("mod", env)

	ampAmp := NewAmp()
	ampAmp.Input("car", amp)
	ampAmp.Input("mod", Value(0.5))

	e.Input("root", ampAmp)

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
	e := &Engine{done: make(chan error)}
	e.sinks("root", &e.root)
	return e
}

type Engine struct {
	sink
	root source

	done chan error
}

func (e *Engine) processAudio(_, out []int16) {
	buf := e.root.Process()
	for i := range buf {
		out[i] = int16(buf[i] * waveAmp)
	}
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
