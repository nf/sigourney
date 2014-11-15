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

// Package debug provides debugging facilities for Sigourney.
package debug

import (
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nf/sigourney/audio"
)

func NewRecorder(p audio.Processor) *Recorder {
	return &Recorder{p: p}
}

type Recorder struct {
	Samples []audio.Sample
	p       audio.Processor
}

func (r *Recorder) Process(s []audio.Sample) {
	r.p.Process(s)
	r.Samples = append(r.Samples, s...)
}

func NewTracer() *Tracer {
	return &Tracer{
		paths: map[string]*Recorder{},
	}
}

type Tracer struct {
	paths map[string]*Recorder
}

func (t *Tracer) Record(label string, p audio.Processor) audio.Processor {
	r := NewRecorder(p)
	t.paths[label] = r
	return r
}

func Process(p audio.Processor, frames int) []audio.Sample {
	out := make([]audio.Sample, frames*audio.FrameLength)
	for i := 0; i < frames; i++ {
		j := i * audio.FrameLength
		p.Process(out[j : j+audio.FrameLength])
	}
	return out
}

const (
	iHeight = 400
	iVScale = 150
)

var iColor = color.Black

func Render(s []audio.Sample) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, len(s), iHeight))
	for x := range s {
		y := iHeight / 2
		dy := y - int(s[x]*iVScale)
		for y != dy && 0 <= y && y < iHeight {
			m.Set(x, y, iColor)
			if y > dy {
				y--
			}
			if y < dy {
				y++
			}
		}
	}
	return m
}

func View(m image.Image) {
	dir, err := ioutil.TempDir("", "sigourney-debug")
	if err != nil {
		panic(err)
	}
	fn := filepath.Join(dir, "image.png")
	f, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	err = png.Encode(f, m)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
	err = exec.Command("open", fn).Run()
	if err != nil {
		panic(err)
	}
}
