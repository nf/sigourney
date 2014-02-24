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
	buf := make([]Sample, nSamples)
	o := NewSin()
	o.Input("pitch", Value(0))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		o.Process(buf)
	}
}

func BenchmarkFMSin(b *testing.B) {
	b.StopTimer()
	buf := make([]Sample, nSamples)
	o, o2 := NewSin(), NewSin()
	o.Input("pitch", o2)
	o2.Input("pitch", Value(0))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		o.Process(buf)
	}
}
