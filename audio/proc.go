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

import (
	"math"

	"github.com/nf/sigourney/fast"
)

func NewOsc() *Osc {
	o := &Osc{}
	o.inputs("pitch", &o.pitch)
	return o
}

type Osc struct {
	sink
	pitch source // 0.1/oct, 0 == 440Hz

	pos float64
}

func (o *Osc) Process(s []Sample) {
	pitch := o.pitch.Process()
	p := o.pos
	for i := range s {
		s[i] = Sample(fast.Sin(p * 2 * math.Pi))
		hz := 440 * math.Exp2(float64(pitch[i])*10)
		p += hz / waveHz
		if p > 100 {
			p -= 100
		}
	}
	o.pos = p
}

func NewAmp() *Amp {
	a := &Amp{}
	a.inputs("car", &a.car, "mod", &a.mod)
	return a
}

type Amp struct {
	sink
	car Processor
	mod source
}

func (a *Amp) Process(s []Sample) {
	a.car.Process(s)
	m := a.mod.Process()
	for i := range s {
		s[i] *= m[i]
	}
}

func NewSum() *Sum {
	s := &Sum{}
	s.inputs("a", &s.a, "b", &s.b)
	return s
}

type Sum struct {
	sink
	a Processor
	b source
}

func (s *Sum) Process(buf []Sample) {
	s.a.Process(buf)
	b := s.b.Process()
	for i := range buf {
		buf[i] += b[i]
	}
}

func NewEnv() *Env {
	e := &Env{}
	e.inputs("att", &e.att, "dec", &e.dec)
	return e
}

type Env struct {
	sink
	att, dec source

	down bool
	v    Sample
}

func (e *Env) Process(s []Sample) {
	att, dec := e.att.Process(), e.dec.Process()
	v := e.v
	for i := range s {
		if e.down {
			if d := dec[i]; d > 0 {
				v -= 1 / (d * waveHz * 10)
			}
		} else {
			if a := att[i]; a > 0 {
				v += 1 / (a * waveHz * 10)
			}
		}
		if v <= 0 {
			v = 0
			e.down = false
		} else if v >= 1 {
			v = 1
			e.down = true
		}
		s[i] = v
	}
	e.v = v
}

type Value Sample

func (v Value) Process(s []Sample) {
	for i := range s {
		s[i] = Sample(v)
	}
}
