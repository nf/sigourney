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

// Dup splits a Processor into multiple audio streams.
type Dup struct {
	src  Processor
	outs []*Output
	buf  []Sample
	done bool
}

func NewDup(src Processor) *Dup {
	d := &Dup{src: src}
	return d
}

func (d *Dup) Tick() {
	d.done = false
}

// SetSource changes the source Processor.
func (d *Dup) SetSource(p Processor) {
	d.src = p
}

// Output creates and returns a new Output Processor.
// Each Output should be Closed when it is no longer in use.
func (d *Dup) Output() *Output {
	o := &Output{d: d}
	d.outs = append(d.outs, o)
	if len(d.outs) > 1 && d.buf == nil {
		d.buf = make([]Sample, nSamples)
	}
	return o
}

// An Output is a Processor endpoint provided by Dup.
type Output struct {
	d *Dup
}

func (o *Output) Process(p []Sample) {
	if !o.d.done {
		o.d.done = true
		o.d.src.Process(p)
		if len(o.d.outs) > 1 {
			copy(o.d.buf, p)
		}
	} else {
		copy(p, o.d.buf)
	}
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
