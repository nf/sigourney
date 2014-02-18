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
	"os"
	"time"

	"github.com/nf/sigourney/audio"
	"github.com/nf/sigourney/ui"
)

func demo() error {
	u, err := ui.New()
	if err != nil {
		return err
	}
	for _, m := range []*ui.Message{
		{Action: "new", Name: "osc1", Kind: "osc"},
		{Action: "new", Name: "osc2", Kind: "osc"},
		{Action: "new", Name: "sum1", Kind: "sum"},
		{Action: "new", Name: "mul1", Kind: "mul"},
		{Action: "new", Name: "mul2", Kind: "mul"},
		{Action: "new", Name: "val1", Kind: "value", Value: 0.1},
		{Action: "connect", From: "osc1", To: "engine", Input: "in"},
		{Action: "connect", From: "mul1", To: "osc1", Input: "pitch"},
		{Action: "connect", From: "osc2", To: "mul1", Input: "a"},
		{Action: "connect", From: "val1", To: "mul1", Input: "b"},
		{Action: "connect", From: "val1", To: "osc2", Input: "pitch"},
	} {
		if err := u.Handle(m); err != nil {
			return err
		}
	}
	time.Sleep(2 * time.Second)
	for _, m := range []*ui.Message{
		{Action: "set", Name: "val1", Value: 0.15},
	} {
		if err := u.Handle(m); err != nil {
			return err
		}
	}
	time.Sleep(2 * time.Second)
	u.Close()

	e := audio.NewEngine()

	oscMod := audio.NewOsc()
	oscMod.Input("pitch", audio.Value(-0.1))

	oscModMul := audio.NewMul()
	oscModMul.Input("a", oscMod)
	oscModMul.Input("b", audio.Value(0.1))

	osc := audio.NewOsc()
	osc.Input("pitch", oscModMul)

	envMod := audio.NewOsc()
	envMod.Input("pitch", audio.Value(-1))

	envModMul := audio.NewMul()
	envModMul.Input("a", envMod)
	envModMul.Input("b", audio.Value(0.02))

	envModSum := audio.NewSum()
	envModSum.Input("a", envModMul)
	envModSum.Input("b", audio.Value(0.021))

	osc2 := audio.NewOsc()
	osc2.Input("pitch", audio.Value(-0.6))

	env := audio.NewEnv()
	env.Input("trig", osc2)
	env.Input("att", audio.Value(0.0001))
	env.Input("dec", envModSum)

	mul := audio.NewMul()
	mul.Input("a", osc)
	mul.Input("b", env)

	mulMul := audio.NewMul()
	mulMul.Input("a", mul)
	mulMul.Input("b", audio.Value(0.5))

	e.Input("in", mulMul)

	if err := e.Start(); err != nil {
		return err
	}

	os.Stdout.Write([]byte("Press enter to stop...\n"))
	os.Stdin.Read([]byte{0})

	return e.Stop()
}
