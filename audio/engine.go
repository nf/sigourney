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

package audio

import "sync"

func NewEngine() *Engine {
	e := &Engine{done: make(chan error)}
	e.inputs("in", &e.in)
	return e
}

type Engine struct {
	sync.Mutex // Held while processing.

	sink
	in source

	done    chan error
	tickers []Ticker
}

func (e *Engine) AddTicker(t Ticker) {
	e.tickers = append(e.tickers, t)
}

func (e *Engine) RemoveTicker(t Ticker) {
	ts := e.tickers
	for i, t2 := range ts {
		if t == t2 {
			copy(ts[i:], ts[i+1:])
			e.tickers = ts[:len(ts)-1]
			break
		}
	}
}

func (e *Engine) Process() []Sample {
	e.Lock()
	buf := e.in.Process()
	for _, t := range e.tickers {
		t.Tick()
	}
	e.Unlock()
	for i := range buf {
		buf[i] *= 0.9 // headroom
	}
	return buf
}

func (e *Engine) Render(frames int) []Sample {
	out := make([]Sample, 0, frames*nSamples)
	for i := 0; i < frames; i++ {
		out = append(out, e.Process()...)
	}
	return out
}
