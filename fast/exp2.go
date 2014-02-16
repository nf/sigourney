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

func Exp2(f float64) float64 {
	i := int((f + exp2TableOffset) * exp2TableFactor)
	if i < 0 || i >= exp2TableLen {
		return math.Exp2(f)
	}
	return exp2Table[i]
}

const (
	exp2TableLen            = 1 << 24 // 128MB table
	exp2Lo, exp2Hi  float64 = -11, 11
	exp2TableOffset         = (exp2Hi - exp2Lo) / 2
	exp2TableFactor         = exp2TableLen / (exp2Hi - exp2Lo)
)

var (
	exp2Table []float64
)

func init() {
	exp2Table = make([]float64, exp2TableLen)
	for i := range exp2Table {
		f := float64(i)/exp2TableFactor - exp2TableOffset
		exp2Table[i] = math.Exp2(f)
	}
}
