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
	env := &Env{attack: waveHz / 100, decay: waveHz / 3}
	e.Track(osc, env)

	amp := &Amp{}
	amp.SetInput("car", osc)
	amp.SetInput("mod", env)
	e.SetInput("", amp)

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
