// +build cgo

/*
Copyright 2014 Google Inc.

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

import "code.google.com/p/portaudio-go/portaudio"

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

func (e *Engine) processAudio(_, out []int16) {
	buf := e.Process()

	// Limiter
	for i, v := range buf {
		if v < 0 {
			v *= -1
		}
		if v > e.max {
			e.max = v
		}
		buf[i] *= 1 / e.max
	}

	for i, s := range buf {
		out[i] = int16(s * waveAmp)
	}
}
