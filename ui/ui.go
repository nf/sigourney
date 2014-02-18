/*
Copyright 2014 Google Inc.

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

package ui

import (
	"errors"
	"log"

	"github.com/nf/sigourney/audio"
)

type Message struct {
	Action string

	// "new", "set", "destroy"
	Name string `json:",omitempty"`

	// "new"
	Kind string `json:",omitempty"`

	// "new", "set"
	Value float64 `json:",omitempty"` // for Kind: "value"

	// "connect", "disconnect"
	From  string `json:",omitEmpty"`
	To    string `json:",omitempty"`
	Input string `json:",omitempty"`

	// "hello"
	ObjectInputs map[string][]string `json:",omitempty"`
}

type UI struct {
	M <-chan *Message
	m chan *Message

	objects map[string]*Object
	engines map[string]*audio.Engine
}

func New() *UI {
	m := make(chan *Message, 1)
	m <- &Message{Action: "hello", ObjectInputs: objectInputs()}
	return &UI{
		M: m, m: m,
		objects: make(map[string]*Object),
		engines: make(map[string]*audio.Engine),
	}
}

func (u *UI) Close() (err error) {
	for _, e := range u.engines {
		if err2 := e.Stop(); err2 != nil && err == nil {
			err = err2
		}
	}
	return
}

func (u *UI) Handle(m *Message) error {
	switch a := m.Action; a {
	case "new":
		o := NewObject(m.Name, m.Kind, m.Value)
		u.objects[m.Name] = o
		if m.Kind == "engine" {
			e := o.proc.(*audio.Engine)
			if err := e.Start(); err != nil {
				return err
			}
			u.engines[m.Name] = e
		}
	case "connect":
		u.connect(m.From, m.To, m.Input)
	case "disconnect":
		u.disconnect(m.From, m.To, m.Input)
	case "set":
		o, ok := u.objects[m.Name]
		if !ok {
			return errors.New("unknown object: " + m.Name)
		}
		v := audio.Value(m.Value)
		o.proc = v
		u.lockEngines()
		o.dup.SetSource(v)
		u.unlockEngines()
	case "destroy":
		o, ok := u.objects[m.Name]
		if !ok {
			return errors.New("bad Name: " + m.Name)
		}
		for d := range o.output {
			u.disconnect(m.Name, d.name, d.input)
		}
		for input, from := range o.Input {
			u.disconnect(from, m.Name, input)
		}
	default:
		log.Println("unrecognized Action:", a)
	}
	return nil
}

func (u *UI) lockEngines() {
	for _, e := range u.engines {
		e.Lock()
	}
}

func (u *UI) unlockEngines() {
	for _, e := range u.engines {
		e.Unlock()
	}
}

func (u *UI) disconnect(from, to, input string) error {
	f, ok := u.objects[from]
	if !ok {
		return errors.New("unknown From: " + from)
	}
	t, ok := u.objects[to]
	if !ok {
		return errors.New("unknown To: " + to)
	}

	u.lockEngines()
	f.output[dest{to, input}].Close()
	t.proc.(audio.Sink).Input(input, audio.Value(0))
	u.unlockEngines()

	delete(f.output, dest{to, input})
	delete(t.Input, input)

	return nil
}

func (u *UI) connect(from, to, input string) error {
	f, ok := u.objects[from]
	if !ok {
		return errors.New("unknown From: " + from)
	}
	t, ok := u.objects[to]
	if !ok {
		return errors.New("unknown To: " + to)
	}

	u.lockEngines()
	o := f.dup.Output()
	t.proc.(audio.Sink).Input(input, o)
	u.unlockEngines()

	f.output[dest{to, input}] = o
	t.Input[input] = to

	return nil
}

type Object struct {
	Name, Kind string
	Input      map[string]string

	proc   interface{}
	dup    *audio.Dup
	output map[dest]*audio.Output
}

type dest struct {
	name, input string
}

func NewObject(name, kind string, value float64) *Object {
	var p interface{}
	switch kind {
	case "osc":
		p = audio.NewOsc()
	case "amp":
		p = audio.NewAmp()
	case "sum":
		p = audio.NewSum()
	case "env":
		p = audio.NewEnv()
	case "engine":
		p = audio.NewEngine()
	case "value":
		p = audio.Value(value)
	default:
		panic("bad kind: " + kind)
	}
	var dup *audio.Dup
	if proc, ok := p.(audio.Processor); ok {
		dup = audio.NewDup(proc)
	}
	return &Object{
		Name: name, Kind: kind,
		Input: make(map[string]string),

		proc:   p,
		dup:    dup,
		output: make(map[dest]*audio.Output),
	}
}

func objectInputs() map[string][]string {
	m := make(map[string][]string)
	for _, k := range kinds {
		o := NewObject("unnamed", k, 0)
		var in []string
		if s, ok := o.proc.(audio.Sink); ok {
			in = s.Inputs()
		}
		m[k] = in
	}
	return m
}

var kinds = []string{"osc", "amp", "sum", "env", "engine", "value"}
