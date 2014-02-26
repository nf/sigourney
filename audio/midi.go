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

import (
	"flag"
	"log"
	"sync"
	"sync/atomic"

	"github.com/rakyll/portmidi"
)

var midiDevice = flag.Int("midi_device", int(portmidi.GetDefaultInputDeviceId()), "MIDI Device ID")

var initMidiOnce sync.Once

func initMidi() {
	s, err := portmidi.NewInputStream(portmidi.DeviceId(*midiDevice), 1024)
	if err != nil {
		log.Println(err)
		return
	}
	go midiLoop(s)
}

var midiNote, midiGate int64 // atomic

func midiLoop(s *portmidi.Stream) {
	var n int64
	for e := range s.Listen() {
		log.Printf("event: %#v\n", e)
		switch e.Status {
		case 144: // note on
			n = e.Data1
			atomic.StoreInt64(&midiNote, n)
			atomic.StoreInt64(&midiGate, 1)
		case 128: // note off
			if e.Data1 == n {
				atomic.StoreInt64(&midiGate, 0)
			}
		}
	}
}

func NewMidiNote() *MidiNote {
	initMidiOnce.Do(initMidi)
	return &MidiNote{}
}

type MidiNote struct{}

func (m *MidiNote) Process(s []Sample) {
	p := (Sample(atomic.LoadInt64(&midiNote)) - 69) / 120
	for i := range s {
		s[i] = p
	}
}

func NewMidiGate() *MidiGate {
	initMidiOnce.Do(initMidi)
	return &MidiGate{}
}

type MidiGate struct{}

func (m *MidiGate) Process(s []Sample) {
	p := Sample(atomic.LoadInt64(&midiGate))
	for i := range s {
		s[i] = p
	}
}
