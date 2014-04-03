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

// Package audio implements the Sigourney audio engine.
package audio

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	nChannels = 1
	nSamples  = 256 * nChannels
)

const (
	waveHz  = 44100
	waveAmp = 1 << 15
)

// A Sample is a single frame of audio.
type Sample float64

// Processor implements an audio source.
//
// As with io.Reader, the caller passes a buffer (or "frame") to the Processor
// and the Processor populates the buffer with sample data.
// Unlike io.Reader, the Processor must populate the entire buffer.
//
// While audio is being generated and the Processor is active, its Process
// method will be called for each audio frame in the stream, so the Processor
// may use on the number of Process calls to maintain its internal state.
type Processor interface {
	Process(buffer []Sample)
}

// A Ticker is a Processor whose Tick method is called once per audio frame.
//
// Each Ticker should be registered with the Engine using AddTicker on
// creation, and similarly removed with RemoveTicker on destruction.
type Ticker interface {
	Tick()
}

// A Sink is a consumer of audio data with one or more named inputs.
type Sink interface {
	// Input attaches the given Processor to the specified named input.
	Input(name string, g Processor)

	// Inputs enumerates the Sink's named inputs.
	Inputs() []string
}

type sink struct {
	m map[string]interface{}
}

func (s *sink) inputs(args ...interface{}) {
	s.m = make(map[string]interface{})
	if len(args)%2 != 0 {
		panic("odd number of args")
	}
	for i := 0; i < len(args); i++ {
		name, ok := args[i].(string)
		if !ok {
			panic("invalid args; expected string")
		}
		i++
		s.m[name] = args[i]

		switch v := args[i].(type) {
		case *Processor:
			*v = Value(0)
		case *source:
			(*v).p = Value(0)
			(*v).b = make([]Sample, nSamples)
		case *trigger:
			(*v).p = Value(0)
			(*v).b = make([]Sample, nSamples)
		case []source:
			for i := range v {
				v[i].p = Value(0)
				v[i].b = make([]Sample, nSamples)
			}
		}
	}
}

func (s *sink) Input(name string, p Processor) {
	if s.m == nil {
		panic("no inputs registered")
	}
	n := strings.Trim(name, "0123456789")
	i, ok := s.m[n]
	if !ok {
		panic("bad input name: " + name)
	}
	switch v := i.(type) {
	case *Processor:
		*v = p
	case *source:
		(*v).p = p
	case *trigger:
		(*v).p = p
	case []source:
		i, _ := strconv.Atoi(strings.TrimPrefix(name, n))
		v[i].p = p
	default:
		panic("bad input type")
	}
}

func (s *sink) Inputs() []string {
	var a []string
	for n, v := range s.m {
		if src, ok := v.([]source); ok {
			for i := range src {
				a = append(a, fmt.Sprint(n, i))
			}
		} else {
			a = append(a, n)
		}
	}
	sort.Strings(a)
	return a
}

type source struct {
	p Processor
	b []Sample
}

func (s *source) Process() []Sample {
	s.p.Process(s.b)
	return s.b
}

const triggerThreshold = 0.5

type trigger struct {
	source
	last bool
}

func (t *trigger) isTrigger(s Sample) bool {
	high := s > triggerThreshold
	trig := !t.last && high
	t.last = high
	return trig

}
