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

import "github.com/nf/gosynth/fix"

const (
	nChannels = 1
	nSamples  = 256 * nChannels
)

const (
	waveHz  = 44100
	waveAmp = 32768
)

type Sample fix.Num

type Processor interface {
	Process([]Sample)
}

type Sink interface {
	Input(name string, g Processor)
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
	}
}

func (s *sink) Input(name string, p Processor) {
	if s.m == nil {
		panic("no inputs registered")
	}
	i, ok := s.m[name]
	if !ok {
		panic("bad input name: " + name)
	}
	switch v := i.(type) {
	case *Processor:
		*v = p
	case *source:
		if (*v).b == nil {
			(*v).b = make([]Sample, nSamples)
		}
		(*v).p = p
	default:
		panic("bad input type")
	}
}

type source struct {
	p Processor
	b []Sample
}

func (s *source) Process() []Sample {
	s.p.Process(s.b)
	return s.b
}
