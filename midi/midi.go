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
	"log"
	"sync"
	"sync/atomic"

	"github.com/rakyll/portmidi"

	"github.com/nf/sigourney/audio"
)

var midiDevice = flag.Int("midi_device", int(portmidi.GetDefaultInputDeviceId()), "MIDI Device ID")

var initOnce sync.Once

func initMidi() {
	s, err := portmidi.NewInputStream(portmidi.DeviceId(*midiDevice), 1024)
	if err != nil {
		log.Println(err)
		return
	}
	if s == nil {
		log.Println("could not initialize MIDI input device")
		return
	}
	go midiLoop(s)
}

var midiNote, midiGate int64 // atomic

func midiLoop(s *portmidi.Stream) {
	noteOn := make([]int64, 0, 128)
	for e := range s.Listen() {
		switch e.Status {
		case 144: // note on
			on := false
			for _, n := range noteOn {
				if n == e.Data1 {
					on = true
				}
			}
			if !on {
				noteOn = append(noteOn, e.Data1)
			}
			atomic.StoreInt64(&midiNote, e.Data1)
			atomic.StoreInt64(&midiGate, 1)
		case 128: // note off
			for i, n := range noteOn {
				if n == e.Data1 {
					copy(noteOn[i:], noteOn[i+1:])
					noteOn = noteOn[:len(noteOn)-1]
				}
			}
			if len(noteOn) > 0 {
				n := noteOn[len(noteOn)-1]
				atomic.StoreInt64(&midiNote, n)
			} else {
				atomic.StoreInt64(&midiGate, 0)
			}
		}
	}
}

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
