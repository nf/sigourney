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
	i := int((f + exp2Offset) * exp2Factor)
	if i < 0 || i >= exp2Len {
		return math.Exp2(f)
	}
	return exp2[i]
}

const (
	exp2Len        = 1 << 24 // 128MB table
	exp2Lo, exp2Hi = -11, 11 // the range of inputs covered by the table
	exp2Offset     = (exp2Hi - exp2Lo) / 2
	exp2Factor     = exp2Len / (exp2Hi - exp2Lo)
)

var exp2 []float64

func init() {
	exp2 = make([]float64, exp2Len)
	for i := range exp2 {
		f := float64(i)/exp2Factor - exp2Offset
		exp2[i] = math.Exp2(f)
	}
}
