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

// Fast Sine approximation with linear interpolation.
func Sin(x float64) float64 {
	if x > 0 {
		return sinLi(x)
	} else {
		return -1 * sinLi(-1*x)
	}
}

func sinLi(x float64) float64 {
	f := x * sinFactor
	t := int(f)
	i := t & (sinLen - 1)
	return sin[i] + grd[i]*(f-float64(t))
}

const (
	sinLen    = 1 << 10 // 1K entry lookup table.
	sinFactor = sinLen / (2 * math.Pi)
)

var sin []float64
var grd []float64

func init() {
	sin = make([]float64, sinLen)
	grd = make([]float64, sinLen)
	step := 1 / sinFactor
	for i := 0; i < sinLen; i++ {
		sin[i] = math.Sin(float64(i) * step)
	}
	for i := 0; i < sinLen-1; i++ {
		grd[i] = sin[i+1] - sin[i]
	}
}
