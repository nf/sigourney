// +build ignore

package main

import (
	"github.com/nf/sigourney/audio"
	"github.com/nf/sigourney/debug"
)

func main() {
	fs := audio.NewSin()
	fs.Input("pitch", audio.Value(-0.4))
	fm := audio.NewMul()
	fm.Input("a", audio.Value(0.2))
	fm.Input("b", fs)

	s := audio.NewSquare()
	s.Input("pitch", audio.Value(-0.1))
	f := audio.NewFilter()
	f.Input("in", s)
	f.Input("freq", fm)

	debug.View(debug.Render(debug.Process(f, 10)))
}
