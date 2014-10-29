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

package protocol

type Message struct {
	Action string

	// Incoming messages

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

	// "setDisplay"
	Display Display `json:",omitempty"`

	// Outgoing messages

	// "hello"
	KindInputs map[string][]string `json:",omitempty"`

	// "setGraph"
	Graph []*Object `json:",omitempty"`

	// "message"
	Message string
}

type Object struct {
	Name    string
	Kind    string
	Value   float64
	Input   map[string]string
	Display Display
}

type Display struct {
	Top   int
	Left  int
	Label string
}
