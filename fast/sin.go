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
		return sinTable[int(f*sinTableFactor*-1)%sinTableLen] * -1
	}
	return sinTable[int(f*sinTableFactor)%sinTableLen]
}

const sinTableLen = 1 << 21 // 16MB table

var (
	sinTable       []float64
	sinTableFactor = sinTableLen / (2 * math.Pi)
)

func init() {
	sinTable = make([]float64, sinTableLen)
	step := 1 / sinTableFactor
	for i := range sinTable {
		sinTable[i] = math.Sin(float64(i) * step)
	}
}
