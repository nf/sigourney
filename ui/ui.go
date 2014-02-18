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
	engine  *audio.Engine
}

func New() (*UI, error) {
	m := make(chan *Message, 1)
	m <- &Message{Action: "hello", ObjectInputs: objectInputs()}
	objs := make(map[string]*Object)
	objs["engine"] = NewObject("engine", "engine", 0)
	e := objs["engine"].proc.(*audio.Engine)
	if err := e.Start(); err != nil {
		return nil, err
	}
	return &UI{M: m, m: m, objects: objs, engine: e}, nil
}

func (u *UI) Close() error {
	return u.engine.Stop()
}

func (u *UI) Handle(m *Message) error {
	switch a := m.Action; a {
	case "new":
		o := NewObject(m.Name, m.Kind, m.Value)
		if o.dup != nil {
			u.engine.Lock()
			u.engine.AddTicker(o.dup)
			u.engine.Unlock()
		}
		u.objects[m.Name] = o
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
		u.engine.Lock()
		o.dup.SetSource(v)
		u.engine.Unlock()
	case "destroy":
		o, ok := u.objects[m.Name]
		if !ok {
			return errors.New("bad Name: " + m.Name)
		}
		if o.dup != nil {
			u.engine.Lock()
			u.engine.RemoveTicker(o.dup)
			u.engine.Unlock()
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

func (u *UI) disconnect(from, to, input string) error {
	f, ok := u.objects[from]
	if !ok {
		return errors.New("unknown From: " + from)
	}
	t, ok := u.objects[to]
	if !ok {
		return errors.New("unknown To: " + to)
	}

	u.engine.Lock()
	f.output[dest{to, input}].Close()
	t.proc.(audio.Sink).Input(input, audio.Value(0))
	u.engine.Unlock()

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

	u.engine.Lock()
	o := f.dup.Output()
	t.proc.(audio.Sink).Input(input, o)
	u.engine.Unlock()

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

var kinds = []string{
	"amp",
	"engine",
	"env",
	"osc",
	"sum",
	"value",
}
