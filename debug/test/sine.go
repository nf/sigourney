// +build ignore

package main

import (
	"github.com/nf/sigourney/audio"
	"github.com/nf/sigourney/debug"
)

func main() {
	s := audio.NewSin()
	s.Input("pitch", audio.Value(-0.3))
	debug.View(debug.Render(debug.Process(s, 5)))
}
