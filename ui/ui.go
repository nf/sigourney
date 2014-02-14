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

	// "new"
	Name  string  `json:",omitempty"`
	Kind  string  `json:",omitempty"`
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
	case "connect", "disconnect":
		to, ok := u.objects[m.To]
		if !ok {
			return errors.New("bad To: " + m.To)
		}
		var from *Object
		if a == "connect" {
			from, ok = u.objects[m.From]
			if !ok {
				return errors.New("bad From: " + m.From)
			}
		}
		u.lockEngines()
		to.SetInput(m.Input, from)
		u.unlockEngines()
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

type Object struct {
	Name, Kind string
	Input      map[string]string

	proc interface{}
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
	return &Object{
		Name: name, Kind: kind,
		Input: make(map[string]string),
		proc:  p,
	}
}

func (o *Object) SetInput(name string, p2 *Object) {
	s := o.proc.(audio.Sink)
	if p2 != nil {
		o.Input[name] = p2.Name
		s.Input(name, p2.proc.(audio.Processor))
	} else {
		delete(o.Input, name)
		s.Input(name, audio.Value(0))
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
