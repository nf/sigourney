package audio

import "testing"

func TestDelay(t *testing.T) {
	sum := NewSum()
	sum.Input("a", Value(1))
	dly := NewDelay()
	dly.Input("in", sum)
	dly.Input("len", Value(Sample(nSamples)/waveHz))
	dup := NewDup(dly)
	out := dup.Output()
	sum.Input("b", dup.Output())

	b := make([]Sample, nSamples)
	for i := 0; i < 6; i++ {
		dup.Tick()
		out.Process(b)
	}
	if b[0] != 3 {
		t.Errorf("b[0] == %v, want 3", b[0])
	}
}
