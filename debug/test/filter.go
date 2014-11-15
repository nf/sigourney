// +build ignore

package main

import (
	"github.com/nf/sigourney/audio"
	"github.com/nf/sigourney/debug"
)

func main() {
	s := audio.NewSquare()
	s.Input("pitch", audio.Value(-0.1))
	f := audio.NewFilter()
	f.Input("in", s)
	debug.View(debug.Render(debug.Process(f, 5)))
}
