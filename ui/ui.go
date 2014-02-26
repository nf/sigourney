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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/nf/sigourney/audio"
)

type Message struct {
	Action string

	// "new", "set", "destroy", "save", "load", "setDisplay"
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
	KindInputs map[string][]string `json:",omitempty"`

	// "setDisplay"
	Display map[string]interface{} `json:",omitempty"`

	// "message"
	Message string
}

type UI struct {
	M <-chan *Message
	m chan *Message

	objects map[string]*Object
	engine  *audio.Engine
}

func New() (*UI, error) {
	m := make(chan *Message, 1)
	u := &UI{M: m, m: m, objects: make(map[string]*Object)}
	u.newObject("engine", "engine", 0)
	u.engine = u.objects["engine"].proc.(*audio.Engine)
	if err := u.engine.Start(); err != nil {
		return nil, err
	}
	m <- &Message{Action: "hello", KindInputs: kindInputs()}
	return u, nil
}

func (u *UI) Close() error {
	return u.engine.Stop()
}

func (u *UI) Handle(m *Message) (err error) {
	defer func() {
		if err != nil {
			u.m <- &Message{
				Action:  "message",
				Message: err.Error(),
			}
		}
	}()
	switch a := m.Action; a {
	case "new":
		u.newObject(m.Name, m.Kind, m.Value)
	case "connect":
		return u.connect(m.From, m.To, m.Input)
	case "disconnect":
		return u.disconnect(m.From, m.To, m.Input)
	case "set":
		o, ok := u.objects[m.Name]
		if !ok {
			return errors.New("unknown object: " + m.Name)
		}
		o.Value = m.Value
		v := audio.Value(m.Value)
		o.proc = v
		u.engine.Lock()
		o.dup.SetSource(v)
		u.engine.Unlock()
	case "destroy":
		return u.destroy(m.Name)
	case "save":
		return u.save(m.Name)
	case "load":
		if err := u.engine.Stop(); err != nil {
			return err
		}
		for name := range u.objects {
			if name != "engine" {
				if err := u.destroy(name); err != nil {
					return err
				}
			}
		}
		if err := u.load(m.Name); err != nil {
			return err
		}
		return u.engine.Start()
	case "setDisplay":
		o, ok := u.objects[m.Name]
		if !ok {
			return errors.New("bad Name: " + m.Name)
		}
		for k, v := range m.Display {
			if o.Display == nil {
				o.Display = make(map[string]interface{})
			}
			o.Display[k] = v
		}
	default:
		return fmt.Errorf("unrecognized Action:", a)
	}
	return nil
}

func (u *UI) destroy(name string) error {
	o, ok := u.objects[name]
	if !ok {
		return errors.New("bad Name: " + name)
	}
	if o.dup != nil {
		u.engine.Lock()
		u.engine.RemoveTicker(o.dup)
		u.engine.Unlock()
	}
	for d := range o.output {
		u.disconnect(name, d.name, d.input)
	}
	for input, from := range o.Input {
		u.disconnect(from, name, input)
	}
	delete(u.objects, name)
	return nil
}

const filePrefix = "patch/"

var validName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func (u *UI) save(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("name %q doesn't match %v", name, validName)
	}
	path := filepath.Join(filePrefix, name)
	b, err := json.MarshalIndent(u.objects, "", "  ")
	if err != nil {
		return fmt.Errorf("save: %v", err)
	}
	return ioutil.WriteFile(path, b, 0644)
}

func (u UI) load(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("load: name doesn't match %v", validName)
	}
	f, err := os.Open(filepath.Join(filePrefix, name))
	if err != nil {
		return fmt.Errorf("load: %v", err)
	}
	defer f.Close()
	objs := make(map[string]*Object)
	if err := json.NewDecoder(f).Decode(&objs); err != nil {
		return fmt.Errorf("load: %v", err)
	}
	for _, o := range objs {
		if o.Kind != "engine" {
			u.newObject(o.Name, o.Kind, float64(o.Value))
		}
		u.m <- &Message{
			Action:  "new",
			Name:    o.Name,
			Kind:    o.Kind,
			Value:   o.Value,
			Display: o.Display,
		}
	}
	for to, o := range objs {
		for input, from := range o.Input {
			if err := u.connect(from, to, input); err != nil {
				return err
			}
			u.m <- &Message{
				Action: "connect",
				From:   from,
				To:     to,
				Input:  input,
			}
		}
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
	t.Input[input] = from

	return nil
}

type Object struct {
	Name    string
	Kind    string
	Value   float64
	Input   map[string]string
	Display map[string]interface{}

	proc   interface{}
	dup    *audio.Dup
	output map[dest]*audio.Output
}

type dest struct {
	name, input string
}

func (u *UI) newObject(name, kind string, value float64) {
	o := NewObject(name, kind, value)
	if o.dup != nil {
		u.engine.Lock()
		u.engine.AddTicker(o.dup)
		u.engine.Unlock()
	}
	u.objects[name] = o
}

func NewObject(name, kind string, value float64) *Object {
	var p interface{}
	switch kind {
	case "clip":
		p = audio.NewClip()
	case "engine":
		p = audio.NewEngine()
	case "env":
		p = audio.NewEnv()
	case "mul":
		p = audio.NewMul()
	case "rand":
		p = audio.NewRand()
	case "sin":
		p = audio.NewSin()
	case "square":
		p = audio.NewSquare()
	case "sum":
		p = audio.NewSum()
	case "value":
		p = audio.Value(value)
	case "note":
		p = audio.NewMidiNote()
	case "gate":
		p = audio.NewMidiGate()
	default:
		panic("bad kind: " + kind)
	}
	var dup *audio.Dup
	if proc, ok := p.(audio.Processor); ok {
		dup = audio.NewDup(proc)
	}
	return &Object{
		Name:  name,
		Kind:  kind,
		Value: value,
		Input: make(map[string]string),

		proc:   p,
		dup:    dup,
		output: make(map[dest]*audio.Output),
	}
}

func kindInputs() map[string][]string {
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
	"clip",
	"engine",
	"env",
	"gate",
	"mul",
	"rand",
	"sin",
	"note",
	"square",
	"sum",
	"value",
}
