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

package fast

import "math"

func Sin(f float64) float64 {
	if f < 0 {
		return sin[int(f*sinFactor*-1)%sinLen] * -1
	}
	return sin[int(f*sinFactor)%sinLen]
}

const (
	sinLen    = 1 << 21 // 16MB table
	sinFactor = sinLen / (2 * math.Pi)
)

var sin []float64

func init() {
	sin = make([]float64, sinLen)
	step := 1 / sinFactor
	for i := range sin {
		sin[i] = math.Sin(float64(i) * step)
	}
}
