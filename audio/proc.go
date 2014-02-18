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
		hz := 440 * fast.Exp2(float64(pitch[i])*10)
		p += hz / waveHz
		if p > 100 {
			p -= 100
		}
	}
	o.pos = p
}

func NewMul() *Mul {
	a := &Mul{}
	a.inputs("a", &a.a, "b", &a.b)
	return a
}

type Mul struct {
	sink
	a Processor
	b source
}

func (a *Mul) Process(s []Sample) {
	a.a.Process(s)
	m := a.b.Process()
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

func NewDup(src Processor) *Dup {
	d := &Dup{src: src}
	return d
}

type Dup struct {
	src  Processor
	outs []*Output
	buf  []Sample
	done bool
}

func (d *Dup) Tick() {
	d.done = false
}

func (d *Dup) SetSource(p Processor) {
	d.src = p
}

func (d *Dup) Output() *Output {
	o := &Output{d: d}
	d.outs = append(d.outs, o)
	if len(d.outs) > 1 && d.buf == nil {
		d.buf = make([]Sample, nSamples)
	}
	return o
}

type Output struct {
	d *Dup
}

func (o *Output) Process(p []Sample) {
	if len(o.d.outs) == 1 {
		if !o.d.done {
			o.d.done = true
			o.d.src.Process(p)
		}
		return
	}
	if !o.d.done {
		o.d.done = true
		o.d.src.Process(o.d.buf)
	}
	copy(p, o.d.buf)
}

func (o *Output) Close() {
	outs := o.d.outs
	for i, o2 := range outs {
		if o == o2 {
			copy(outs[i:], outs[i+1:])
			o.d.outs = outs[:len(outs)-1]
			break
		}
	}
}
