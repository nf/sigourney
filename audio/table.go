/*
Copyright 2015 Google Inc.

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

import "math"

type TableOsc struct {
	sink
	table []float64
	pitch Processor
	syn   trigger

	pos float64
}

func NewTableOsc(table []float64) *TableOsc {
	w := &TableOsc{table: table}
	w.inputs("pitch", &w.pitch, "syn", &w.syn)
	return w
}

func (w *TableOsc) Process(s []Sample) {
	p := w.pos
	w.pitch.Process(s)
	t := w.syn.Process()
	hz, lastS := sampleToHz(s[0]), s[0]
	for i := range s {
		if w.syn.isTrigger(t[i]) {
			p = 0
		}
		if s[i] != lastS {
			hz = sampleToHz(s[i])
		}
		s[i] = Sample(w.table[int(p)])
		p += hz / waveHz * float64(len(w.table))
		for p > float64(len(w.table)-1) {
			p -= float64(len(w.table))
		}
	}
	w.pos = p
}

var (
	bandLimitedSquareTable   []float64
	bandLimitedTriangleTable []float64
	bandLimitedSawTable      []float64
)

func init() {
	oddHarmonics := []int{1, 3, 5, 7, 9, 11}
	allHarmonics := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	nSamples := 1024 * 16
	bandLimitedSquareTable = newHarmonicTable(nSamples, oddHarmonics, func(k int) float64 {
		return 1 / float64(k)
	})
	bandLimitedTriangleTable = newHarmonicTable(nSamples, oddHarmonics, func(k int) float64 {
		return 1 / float64(k) / float64(k)
	})
	bandLimitedSawTable = newHarmonicTable(nSamples, allHarmonics, func(k int) float64 {
		return 2. / math.Pi * math.Pow(-1.0, float64(k)) / float64(k)
	})
}

func newHarmonicTable(samples int, harmonics []int, amp func(int) float64) []float64 {
	table := make([]float64, samples)
	max := 0.0
	for i := range table {
		pos := float64(i) / float64(len(table))
		for _, h := range harmonics {
			a := 1.0 * amp(h)
			freq := 2 * math.Pi * float64(h)
			table[i] += a * math.Sin(freq*pos)
		}
		if table[i] > max {
			max = table[i]
		}
	}
	for i := range table {
		table[i] /= max
	}
	return table
}

func NewBandLimitedSquare() *TableOsc   { return NewTableOsc(bandLimitedSquareTable) }
func NewBandLimitedTriangle() *TableOsc { return NewTableOsc(bandLimitedTriangleTable) }
func NewBandLimitedSaw() *TableOsc      { return NewTableOsc(bandLimitedSawTable) }
