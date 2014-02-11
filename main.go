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

	osc := &SimpleOsc{}
	e.Track(osc)

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
	buf     []Sample
	tickers []Ticker
	root    Processor
	done    chan error
}

func (e *Engine) Track(t ...Ticker) {
	e.tickers = append(e.tickers, t...)
}

func (e *Engine) Tick() {
	for _, t := range e.tickers {
		t.Tick()
	}
}

func (e *Engine) ProcessAudio(_, out []int16) {
	e.root.Process(e.buf)
	for i := range e.buf {
		out[i] = int16(e.buf[i] * waveAmp)
	}
	e.Tick()
}

func (e *Engine) SetInput(_ string, p Processor) {
	e.root = p
}

func (e *Engine) Start() error {
	stream, err := portaudio.OpenDefaultStream(0, 1, waveHz, nSamples, e.ProcessAudio)
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
