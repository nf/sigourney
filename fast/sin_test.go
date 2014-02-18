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

func TestSin(t *testing.T) {
	const accuracy = 0.0001
	for _, f := range []float64{
		math.Pi, 10000 * math.Pi, 2 * math.Pi, 1000 * 2 * math.Pi,
		0, 1, 0.5, -1, -0.5, -1000, 1000,
		-9.6466616586009283766724726e-01,
		9.9338225271646545763467022e-01,
		-2.7335587039794393342449301e-01,
		9.5586257685042792878173752e-01,
		-2.099421066779969164496634e-01,
		2.135578780799860532750616e-01,
		-8.694568971167362743327708e-01,
		4.019566681155577786649878e-01,
		9.6778633541687993721617774e-01,
		-6.734405869050344734943028e-01,
	} {
		got, want := Sin(f), math.Sin(f)
		diff := math.Abs(got - want)
		if diff > accuracy {
			t.Errorf("Sin(%v) = %v, want = %v ± %v", f, got, want, accuracy)
		}
	}
}
