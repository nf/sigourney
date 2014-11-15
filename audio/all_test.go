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

import "testing"

func BenchmarkSin(b *testing.B) {
	b.StopTimer()
	buf := make([]Sample, FrameLength)
	o := NewSin()
	o.Input("pitch", Value(0))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		o.Process(buf)
	}
}

func BenchmarkFMSin(b *testing.B) {
	b.StopTimer()
	buf := make([]Sample, FrameLength)
	o, o2 := NewSin(), NewSin()
	o.Input("pitch", o2)
	o2.Input("pitch", Value(0))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		o.Process(buf)
	}
}

func TestDelay(t *testing.T) {
	sum := NewSum()
	sum.Input("a", Value(1))
	dly := NewDelay()
	dly.Input("in", sum)
	dly.Input("len", Value(Sample(FrameLength)/waveHz))
	dup := NewDup(dly)
	out := dup.Output()
	sum.Input("b", dup.Output())

	b := make([]Sample, FrameLength)
	for i := 0; i < 6; i++ {
		dup.Tick()
		out.Process(b)
		t.Logf("iteration %d: b[0] == %v", i, b[0])
	}
	if b[0] != 3 {
		t.Errorf("b[0] == %v, want 3", b[0])
	}
}

func TestDup(t *testing.T) {
	var p countingProcessor
	d := NewDup(&p)
	o1, o2 := d.Output(), d.Output()

	n := 0
	b := make([]Sample, FrameLength)

	check := func() {
		if int(p) != n {
			t.Errorf("n == %v, p == %v", n, p)
		}
		if int(b[0]) != n {
			t.Errorf("n == %v, b[0] == %v", n, b[0])
		}
	}
	tick := func() {
		n++
		d.Tick()
	}

	tick()
	o1.Process(b)
	check()
	o2.Process(b)
	check()

	tick()
	o1.Process(b)
	check()

	tick()
	o1.Process(b)
	check()
	o2.Process(b)
	check()

	tick()
	o2.Process(b)
	check()
	o1.Process(b)
	check()

	tick()
	o2.Process(b)
	check()

	tick()
	o2.Process(b)
	check()

	o3 := d.Output()

	tick()
	o3.Process(b)
	check()
	o2.Process(b)
	check()
	o1.Process(b)
	check()

	tick()
	o1.Process(b)
	check()
	o3.Process(b)
	check()

	tick()
	o3.Process(b)
	check()
	o1.Process(b)
	check()

	o4 := d.Output()

	tick()
	o3.Process(b)
	check()
	o4.Process(b)
	check()

	tick()
	o4.Process(b)
	check()
	o4.Process(b)
	check()
	o4.Process(b)
	check()
	o4.Process(b)
	check()
	o4.Process(b)
	check()
}

type countingProcessor int

func (p *countingProcessor) Process(b []Sample) {
	(*p)++
	Value(*p).Process(b)
}
