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

// Fast Sine approximation with table lookup and linear interpolation.
func Sin(x float64) float64 {
	f := x * sinFactor
	if x < 0 {
		f *= -1
	}
	t := int(f)
	i := t & (sinLen - 1)
	res := sinTable[i] + sinGrad[i]*(f-float64(t))
	if x < 0 {
		return res * -1
	}
	return res
}

const (
	sinLen    = 512
	sinFactor = sinLen / (2 * math.Pi)
)

var sinTable, sinGrad []float64

func init() {
	sinTable = make([]float64, sinLen)
	sinGrad = make([]float64, sinLen)
	step := 1 / sinFactor
	for i := 0; i < sinLen; i++ {
		sinTable[i] = math.Sin(float64(i) * step)
	}
	for i := 0; i < sinLen; i++ {
		sinGrad[i] = sinTable[(i+1)%sinLen] - sinTable[i]
	}
}
