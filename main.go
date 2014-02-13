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

package main

import (
	"log"
	"os"

	"code.google.com/p/portaudio-go/portaudio"
	"github.com/nf/gosynth/audio"
)

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	if err := demo(); err != nil {
		log.Println(err)
	}
}

func demo() error {
	e := audio.NewEngine()

	oscMod := audio.NewOsc()
	oscMod.Input("pitch", audio.Value(-0.1))

	oscModAmp := audio.NewAmp()
	oscModAmp.Input("car", oscMod)
	oscModAmp.Input("mod", audio.Value(0.1))

	osc := audio.NewOsc()
	osc.Input("pitch", oscModAmp)

	envMod := audio.NewOsc()
	envMod.Input("pitch", audio.Value(-1))

	envModAmp := audio.NewAmp()
	envModAmp.Input("car", envMod)
	envModAmp.Input("mod", audio.Value(0.02))

	envModSum := audio.NewSum()
	envModSum.Input("a", envModAmp)
	envModSum.Input("b", audio.Value(0.021))

	env := audio.NewEnv()
	env.Input("att", audio.Value(0.0001))
	env.Input("dec", envModSum)

	amp := audio.NewAmp()
	amp.Input("car", osc)
	amp.Input("mod", env)

	ampAmp := audio.NewAmp()
	ampAmp.Input("car", amp)
	ampAmp.Input("mod", audio.Value(0.5))

	e.Input("root", ampAmp)

	if err := e.Start(); err != nil {
		return err
	}

	os.Stdout.Write([]byte("Press enter to stop...\n"))
	os.Stdin.Read([]byte{0})

	return e.Stop()
}
