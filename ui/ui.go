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

type Handler interface {
	Hello(kindInputs map[string][]string)
	New(o *Object)
	Connect(from, to, input string)
}

type UI struct {
	h Handler

	objects map[string]*Object
	engine  *audio.Engine
}

func New(h Handler) (*UI, error) {
	u := &UI{h: h, objects: make(map[string]*Object)}
	u.NewObject("engine", "engine", 0)
	u.engine = u.objects["engine"].proc.(*audio.Engine)
	if err := u.engine.Start(); err != nil {
		return nil, err
	}
	h.Hello(kindInputs())
	return u, nil
}

func (u *UI) Close() error {
	return u.engine.Stop()
}

func (u *UI) Destroy(name string) error {
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
		u.Disconnect(name, d.name, d.input)
	}
	for input, from := range o.Input {
		u.Disconnect(from, name, input)
	}
	delete(u.objects, name)
	return nil
}

const filePrefix = "patch/"

var validName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func (u *UI) Save(name string) error {
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

func (u *UI) Load(name string) error {
	if err := u.engine.Stop(); err != nil {
		return err
	}
	for name := range u.objects {
		if name != "engine" {
			if err := u.Destroy(name); err != nil {
				return err
			}
		}
	}
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
			u.NewObject(o.Name, o.Kind, float64(o.Value))
		}
		u.objects[o.Name].Display = o.Display
		u.h.New(o)
	}
	for to, o := range objs {
		for input, from := range o.Input {
			if err := u.Connect(from, to, input); err != nil {
				return err
			}
			u.h.Connect(from, to, input)
		}
	}
	return u.engine.Start()
}

func (u *UI) Disconnect(from, to, input string) error {
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

func (u *UI) Connect(from, to, input string) error {
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

func (u *UI) Set(name string, v float64) error {
	o, ok := u.objects[name]
	if !ok {
		return errors.New("unknown object: " + name)
	}
	o.Value = v
	av := audio.Value(v)
	o.proc = av
	u.engine.Lock()
	o.dup.SetSource(av)
	u.engine.Unlock()
	return nil
}

func (u *UI) SetDisplay(name string, display map[string]interface{}) error {
	o, ok := u.objects[name]
	if !ok {
		return errors.New("unknown object: " + name)
	}
	for k, v := range display {
		if o.Display == nil {
			o.Display = make(map[string]interface{})
		}
		o.Display[k] = v
	}
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

func (u *UI) NewObject(name, kind string, value float64) {
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
	case "delay":
		p = audio.NewDelay()
	case "engine":
		p = audio.NewEngine()
	case "env":
		p = audio.NewEnv()
	case "gate":
		p = audio.NewMidiGate()
	case "mul":
		p = audio.NewMul()
	case "note":
		p = audio.NewMidiNote()
	case "quant":
		p = audio.NewQuant()
	case "rand":
		p = audio.NewRand()
	case "sin":
		p = audio.NewSin()
	case "skip":
		p = audio.NewSkip()
	case "sequencer":
		p = audio.NewStep()
	case "square":
		p = audio.NewSquare()
	case "sum":
		p = audio.NewSum()
	case "value":
		p = audio.Value(value)
	case "filter":
		p = audio.NewFilter()
	case "pole":
		p = audio.NewPole()
	case "noise":
		p = audio.NewNoise()
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
	"delay",
	"engine",
	"env",
	"gate",
	"mul",
	"note",
	"quant",
	"rand",
	"sin",
	"skip",
	"sequencer",
	"square",
	"sum",
	"value",
	"filter",
	"pole",
	"noise",
}
