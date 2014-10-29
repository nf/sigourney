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

package socket

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/gorilla/websocket"

	"github.com/nf/sigourney/protocol"
	"github.com/nf/sigourney/ui"
)

const filePrefix = "patch/"

var validName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func Handler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		log.Println(err)
		return
	}

	s, err := NewSession()
	if err != nil {
		log.Println(err)
		return
	}
	defer s.Close()

	go func() {
		for m := range s.M {
			if err := c.WriteJSON(m); err != nil {
				if err != io.EOF {
					log.Println(err)
				}
				return
			}
		}
	}()

	for {
		m := new(protocol.Message)
		if err := c.ReadJSON(m); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}
		if err := s.Handle(m); err != nil {
			log.Println(err)
		}
	}
}

func NewSession() (*Session, error) {
	m := make(chan *protocol.Message, 1)
	s := &Session{M: m, m: m}
	u := ui.New(s)
	if err := u.Start(); err != nil {
		return nil, err
	}
	s.u = u
	return s, nil
}

type Session struct {
	M <-chan *protocol.Message

	m chan *protocol.Message
	u *ui.UI
}

func (s *Session) Close() error {
	return s.u.Stop()
}

func (s *Session) Hello(kindInputs map[string][]string) {
	s.m <- &protocol.Message{Action: "hello", KindInputs: kindInputs}
}

func (s *Session) SetGraph(graph []*protocol.Object) {
	s.m <- &protocol.Message{Action: "setGraph", Graph: graph}
}

func (s *Session) Handle(m *protocol.Message) (err error) {
	defer func() {
		if err != nil {
			s.m <- &protocol.Message{
				Action:  "message",
				Message: err.Error(),
			}
		}
	}()
	switch a := m.Action; a {
	case "new":
		s.u.NewObject(m.Name, m.Kind, m.Value)
	case "connect":
		return s.u.Connect(m.From, m.To, m.Input)
	case "disconnect":
		return s.u.Disconnect(m.From, m.To, m.Input)
	case "set":
		return s.u.Set(m.Name, m.Value)
	case "destroy":
		return s.u.Destroy(m.Name)
	case "load", "save":
		if !validName.MatchString(m.Name) {
			return fmt.Errorf("name %q doesn't match %v", m.Name, validName)
		}
		filename := filepath.Join(filePrefix, m.Name)
		switch a {
		case "load":
			if err := s.u.Stop(); err != nil {
				return err
			}
			if err := s.u.Load(filename); err != nil {
				return err
			}
			return s.u.Start()
		case "save":
			return s.u.Save(filename)
		}
	case "setDisplay":
		return s.u.SetDisplay(m.Name, m.Display)
	default:
		return fmt.Errorf("unrecognized Action: %v", a)
	}
	return nil
}
