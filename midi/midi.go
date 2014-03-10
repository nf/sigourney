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

package midi

import (
	"flag"
	"sync"
	"sync/atomic"

	"github.com/nf/sigourney/audio"
)

var midiDevice = flag.Int("midi_device", -1, "MIDI Device ID")

var initOnce sync.Once

var midiNote, midiGate int64 // set atomically by midiLoop

func NewNote() *Note {
	initOnce.Do(initMidi)
	return &Note{}
}

type Note struct{}

func (m *Note) Process(s []audio.Sample) {
	p := (audio.Sample(atomic.LoadInt64(&midiNote)) - 69) / 120
	for i := range s {
		s[i] = p
	}
}

func NewGate() *Gate {
	initOnce.Do(initMidi)
	return &Gate{}
}

type Gate struct{}

func (m *Gate) Process(s []audio.Sample) {
	p := audio.Sample(atomic.LoadInt64(&midiGate))
	for i := range s {
		s[i] = p
	}
}
