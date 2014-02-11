package fix

import "testing"

func TestNum(t *testing.T) {
	a := Float(1.5)
	b := Float(6.2)
	t.Log(a + b)
	t.Log(a.Mul(b))
	t.Log(a.Mul(Float(0.001)))
	t.Log(a.Div(Int(-1000)))
	t.Log(Int(1).Div(Int(1000000)))
}
