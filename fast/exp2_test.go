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

import (
	"math"
	"testing"
)

func TestExp2(t *testing.T) {
	const accuracy = 0.0001
	for _, f := range []float64{
		7, 6, 5, 4, 3, 2, 1, 0.5, 0,
		-10, -9, -8, -7, -6 - 5, -4, -3, -2, -1, -0.5,
		3.1537839463286288034313104e+01,
		2.1361549283756232296144849e+02,
		8.2537402562185562902577219e-01,
		3.1021158628740294833424229e-02,
		7.9581744110252191462569661e+02,
		7.6019905892596359262696423e+00,
		3.7506882048388096973183084e+01,
		6.6250893439173561733216375e+00,
		3.5438267900243941544605339e+00,
		2.4281533133513300984289196e-03,
	} {
		got, want := Exp2(f), math.Exp2(f)
		delta := math.Abs(got - want)
		if delta > accuracy {
			t.Errorf("Exp2(%v) = %v, want = %v ± %v (delta %v)", f, got, want, accuracy, delta)
		}
	}
}
