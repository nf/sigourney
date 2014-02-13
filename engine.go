/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"sync"

	"code.google.com/p/portaudio-go/portaudio"
)

func NewEngine() *Engine {
	e := &Engine{done: make(chan error)}
	e.inputs("root", &e.root)
	return e
}

type Engine struct {
	sync.Mutex // Held while processing.

	sink
	root source

	done chan error
}

func (e *Engine) processAudio(_, out []int16) {
	e.Lock()
	buf := e.root.Process()
	e.Unlock()
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
