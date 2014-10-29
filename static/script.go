// build +js

package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
	"github.com/nf/sigourney/protocol"
)

var Debug = true
var ui *UI

type UI struct {
	kindInputs       map[string][]string
	objects          map[string]*Object
	changedSinceSave bool
	ws               js.Object
	plumb            js.Object
}

type Object struct {
	protocol.Object
	el js.Object
}

func main() {
	ui = &UI{
		kindInputs:       make(map[string][]string),
		objects:          make(map[string]*Object),
		changedSinceSave: false,
	}

	status := jQuery("#status")
	js.Global.Get("jsPlumb").Call("bind", "ready", func() {
		status.Call("text", "Connecting to back end...")
		ui.ws = js.Global.Get("WebSocket").New("ws://localhost:8080/socket")
		ui.ws.Set("onopen", func() {
			status.Call("empty")

			initPlumb()
			initUI()
		})
		ui.ws.Set("onclose", func() {
			status.Call("append", "<div>Lost connection to back end!</div>")
		})
		ui.ws.Set("onmessage", onMessage)
	})
}

func initPlumb() {
	ui.plumb = js.Global.Get("jsPlumb").Call("getInstance", js.M{"Container": "page"})
	ui.plumb.Call("bind", "connection", func(conn js.Object) {
		source := ui.objects[conn.Get("sourceId").Str()]
		target := ui.objects[conn.Get("targetId").Str()]
		input := conn.Get("targetEndpoint").Call("getParameter", "input").Str()
		if target.Input[input] != source.Name {
			target.Input[input] = source.Name
			ui.send(&protocol.Message{Action: "connect", From: source.Name, To: target.Name, Input: input})
			ui.changedSinceSave = true
		}
	})
	ui.plumb.Call("bind", "connectionDetached", func(conn js.Object) {
		if !conn.Get("targetEndpoint").Get("isTarget").Bool() {
			return
		}
		source := ui.objects[conn.Get("sourceId").Str()]
		target := ui.objects[conn.Get("targetId").Str()]
		input := conn.Get("targetEndpoint").Call("getParameter", "input").Str()
		target.Input[input] = ""
		ui.send(&protocol.Message{Action: "disconnect", From: source.Name, To: target.Name, Input: input})
		ui.changedSinceSave = true
	})
	ui.plumb.Call("bind", "click", func(conn, e js.Object) {
		if !e.Get("shiftKey").Bool() {
			return
		}
		ui.plumb.Call("detach", conn)
	})
}

func initUI() {
	fn := jQuery(`<input type="text"/>`)
	load := jQuery(`<input type="button" value="load"/>`)
	save := jQuery(`<input type="button" value="save"/>`)
	jQuery("#control").Call("append", fn, load, save)

	loadFn := func() {
		changeWarning := "There are unsaved changes!\nOK to continue?"
		if ui.changedSinceSave && !js.Global.Call("confirm", changeWarning).Bool() {
			return
		}
		for _, obj := range ui.objects {
			obj.destroy()
		}
		ui.objects = make(map[string]*Object)
		ui.send(&protocol.Message{Action: "load", Name: fn.Call("val").Str()})
		fn.Call("blur")
		load.Call("blur")
		ui.changedSinceSave = false
	}
	fn.Call("keypress", func(e js.Object) {
		if e.Get("charCode").Int() == 13 {
			loadFn()
		}
	})
	load.Call("click", loadFn)
	save.Call("click", func() {
		ui.send(&protocol.Message{Action: "save", Name: fn.Call("val").Str()})
		save.Call("blur")
		ui.changedSinceSave = false
	})

	// Handle keypresses while fn field is not focused.
	jQuery(js.Global.Get("document")).Call("keypress", func(e js.Object) {
		if fn.Call("is", ":focus").Bool() {
			return
		}
		switch e.Get("charCode").Int() {
		case 100: // d
			onDup()
		case 24: // ^x
			onDelete()
		default:
			return
		}
		e.Call("preventDefault")
	})

	// Blur inputs on clicks outside controls.
	jQuery("#page, #objects").Call("mousedown", func(e js.Object) {
		if !jQuery(e.Get("originalEvent").Get("target")).Call("is", "#control input").Bool() {
			jQuery("#control input").Call("blur")
		}
	})

	jQuery("#page").Call("selectable", js.M{"filter": ".object"})
}

func onMessage(msg js.Object) {
	var m protocol.Message
	err := json.Unmarshal([]byte(msg.Get("data").Str()), &m)
	if err != nil {
		panic(err)
	}

	if Debug {
		fmt.Printf("< %+v\n", &m)
	}
	switch m.Action {
	case "hello":
		handleHello(m.KindInputs)
	case "setGraph":
		ui.plumb.Call("doWhileSuspended", func() {
			handleSetGraph(m.Graph)
		})
	case "message":
		div := jQuery("<div></div>")
		div.Call("text", m.Message)
		jQuery("#status").Call("append", div)
		js.Global.Call("setTimeout", func() {
			div.Call("remove")
		}, 5000)
	}
}

func handleHello(inputs map[string][]string) {
	for k, kInputs := range inputs {
		ui.kindInputs[k] = kInputs
		if k == "engine" {
			var d protocol.Display
			w := jQuery("#page").Call("width").Int()
			h := jQuery("#page").Call("height").Int()
			d.Top = 3 * h / 4
			d.Left = w/2 - (49+20+20+1+1)/2
			createObject(k, d)
		} else {
			addKind(k, kInputs)
		}
	}
}

func addKind(kind string, inputs []string) {
	li := jQuery("<li></li>")
	li.Call("text", kind)
	li.Call("data", "inputs", inputs)
	li.Call("addClass", "kind-"+kind)
	li.Call("appendTo", "#objects")
	li.Call("draggable", js.M{
		"revert":         true,
		"revertDuration": 0,
		"helper":         "clone",
		"stop": func(e, jui js.Object) {
			var d protocol.Display
			d.Top = jui.Get("position").Get("top").Int()
			d.Left = jui.Get("position").Get("left").Int()
			createObject(kind, d)
			ui.changedSinceSave = true
		},
	})
}

func handleSetGraph(graph []*protocol.Object) {
	for _, o := range graph {
		bumpNCount(o.Name)
		newObject(o)
	}
	for _, o := range graph {
		for input, from := range o.Input {
			ui.plumb.Call("connect", js.M{"uuids": js.S{from + "-out", o.Name + "-" + input}})
		}
	}
}

var nCount = 0

var nameRegexp = initNameRegexp()

func initNameRegexp() *regexp.Regexp {
	r, err := regexp.Compile("[0-9]+$")
	if err != nil {
		panic(err)
	}
	return r
}

func bumpNCount(name string) {
	r := nameRegexp.FindString(name)
	if r == "" {
		return // unknown name
	}
	n, _ := strconv.Atoi(r)
	if nCount <= n {
		nCount = n + 1
	}
}

func createObject(kind string, display protocol.Display) *Object {
	nCount++
	name := kind + strconv.Itoa(nCount)
	if kind == "engine" {
		name = "engine"
	}

	obj := newObject(&protocol.Object{Name: name, Kind: kind, Display: display})

	if kind != "engine" {
		m := &protocol.Message{Action: "new", Name: name, Kind: kind}
		if kind == "value" {
			m.Value = obj.Value
		}
		ui.send(m)
	}
	ui.onDisplayUpdate(obj)

	return obj
}

func newObject(b *protocol.Object) *Object {
	inputs := make(map[string]string)
	if kInputs, ok := ui.kindInputs[b.Kind]; ok {
		for _, input := range kInputs {
			inputs[input] = ""
		}
	}

	obj := &Object{
		el: nil,
		Object: protocol.Object{
			Name:    b.Name,
			Kind:    b.Kind,
			Value:   b.Value,
			Display: b.Display,
		},
	}
	obj.Input = inputs

	ui.objects[b.Name] = obj
	obj.element()

	return obj
}

func elToObj(el js.Object) *Object {
	return ui.objects[jQuery(el).Call("attr", "id").Str()]
}

func onDup() {
	names := make(map[string]string)
	objs := jQuery(".ui-selected").Call("not", "#engine")
	objs.Call("each", func(_, el js.Object) {
		// duplicate objects
		obj1 := elToObj(el)
		d := obj1.Display
		d.Left += 50
		d.Top += 50
		obj2 := createObject(obj1.Kind, d)
		names[obj1.Name] = obj2.Name
		obj2.element().Call("addClass", "ui-selected")
		if obj1.Kind == "value" {
			obj2.setValue(obj1.Value)
			ui.onSetValue(obj2)
		}
	})
	objs.Call("each", func(_, el js.Object) {
		// connect new objects
		obj := elToObj(el)
		for input := range obj.Input {
			targetName := names[obj.Name]
			sourceName, ok := names[obj.Input[input]]
			if !ok {
				continue
			}
			ui.plumb.Call("connect", js.M{"uuids": js.S{sourceName + "-out", targetName + "-" + input}})
		}
	})
	objs.Call("removeClass", "ui-selected")
}

func onDelete() {
	jQuery(".ui-selected").Call("not", "#engine").Call("each", func(_, el js.Object) {
		obj := elToObj(el)
		obj.destroy()
		ui.onDestroy(obj)
	})
}

func (ui *UI) send(m *protocol.Message) {
	if Debug {
		fmt.Printf("> %+v\n", m)
	}
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	ui.ws.Call("send", string(data))
}

func (ui *UI) onDisplayUpdate(obj *Object) {
	ui.send(&protocol.Message{Action: "setDisplay", Name: obj.Name, Display: obj.Display})
	ui.changedSinceSave = true
}

func (ui *UI) onSetValue(obj *Object) {
	ui.send(&protocol.Message{Action: "set", Name: obj.Name, Value: obj.Value})
	ui.changedSinceSave = true
}

func (ui *UI) onDestroy(obj *Object) {
	ui.send(&protocol.Message{Action: "destroy", Name: obj.Name})
	ui.changedSinceSave = true
	delete(ui.objects, obj.Name)
}

var endpointCommon = js.M{
	"endpoint": "Dot",
	"paintStyle": js.M{
		"fillStyle": "#DDD",
		"radius":    6,
	},
	"hoverPaintStyle": js.M{
		"fillStyle": "#FFF",
	},
	"connector": js.S{"Flowchart"},
	"connectorStyle": js.M{
		"lineWidth":   3,
		"strokeStyle": "#BBB",
	},
	"connectorHoverStyle": js.M{
		"strokeStyle": "#FFF",
	},
}

func (obj *Object) element() js.Object {
	if obj.el != nil {
		return obj.el
	}

	el := jQuery(`<div class="object"></div>`)
	el.Call("attr", "id", obj.Name)
	el.Call("text", obj.Kind)
	el.Call("addClass", "kind-"+obj.Kind)
	el.Call("css", "top", obj.Display.Top)
	el.Call("css", "left", obj.Display.Left)
	el.Call("appendTo", "#page")
	obj.el = el

	if obj.Display.Label != "" {
		obj.el.Call("text", obj.Display.Label)
	}

	ui.plumb.Call("draggable", obj.el, js.M{
		"drag": func(e, jqUI js.Object) {
			if !obj.el.Call("is", ".ui-selected").Bool() {
				return
			}
			o1 := obj.el.Call("offset")
			p := jqUI.Get("position")
			jQuery(".ui-selected").Call("not", obj.el).Call("each", func(_, el js.Object) {
				o2 := jQuery(el).Call("offset")
				jQuery(el).Call("css", js.M{
					"top":  p.Get("top").Int() - o1.Get("top").Int() + o2.Get("top").Int(),
					"left": p.Get("left").Int() - o1.Get("left").Int() + o2.Get("left").Int(),
				})
				ui.plumb.Call("repaint", el)
			})
		},
		"stop": func() {
			obj.updateOffset()
			ui.onDisplayUpdate(obj)
			if !obj.el.Call("is", ".ui-selected").Bool() {
				return
			}
			jQuery(".ui-selected").Call("not", obj.el).Call("each", func(_, el js.Object) {
				ui.plumb.Call("repaint", el)
				obj := elToObj(el)
				obj.updateOffset()
				ui.onDisplayUpdate(obj)
			})
		},
	})

	if obj.Kind == "value" {
		obj.setValue(obj.Value)
		obj.el.Call("dblclick", func(e js.Object) {
			s := js.Global.Call("prompt", "Value? (-1 to +1)").Str()
			if n, ok := noteToValue(s); ok {
				obj.setLabel(s)
				obj.setValue(n)
				ui.onDisplayUpdate(obj)
				ui.onSetValue(obj)
				return
			}
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return
			}
			obj.setValue(f)
			ui.onSetValue(obj)
		})
	}

	if obj.Kind != "engine" {
		obj.el.Call("click", func(e js.Object) {
			if !e.Get("shiftKey").Bool() {
				return
			}
			obj.destroy()
			ui.onDestroy(obj)
		})
	}

	ui.plumb.Call("doWhileSuspended", func() {
		for input := range obj.Input {
			ui.plumb.Call("addEndpoint", obj.el, js.M{
				"uuid":       obj.Name + "-" + input,
				"parameters": js.M{"input": input},
				"anchor":     "ContinuousTop",
				"isSource":   false,
				"isTarget":   true,
				"overlays": js.S{
					js.S{
						"Label",
						js.M{
							"label":    input,
							"cssClass": "label",
						},
					},
				},
			}, endpointCommon)
		}

		if obj.Kind != "engine" {
			ui.plumb.Call("addEndpoint", obj.el, js.M{
				"uuid":           obj.Name + "-out",
				"anchor":         "Bottom",
				"isSource":       true,
				"isTarget":       false,
				"maxConnections": -1,
			}, endpointCommon)
		}
	})

	return obj.el
}

func (obj *Object) updateOffset() {
	offset := obj.element().Call("offset")
	obj.Display.Left = offset.Get("left").Int()
	obj.Display.Top = offset.Get("top").Int()
}

func (obj *Object) setValue(v float64) {
	obj.Value = v
	if obj.Display.Label == "" {
		ui.plumb.Call("repaint", obj.element().Call("text", v))
	}
}

func (obj *Object) setLabel(l string) {
	obj.Display.Label = l
	ui.plumb.Call("repaint", obj.element().Call("text", l))
}

func (obj *Object) destroy() {
	if obj.el.IsNull() {
		return
	}
	ui.plumb.Call("remove", jQuery(obj.el))
}

var noteRegexp = initNoteRegexp()
var tones = map[string]int{"c": -9, "d": -7, "e": -5, "f": -4, "g": -2, "a": 0, "b": 2}

func initNoteRegexp() *regexp.Regexp {
	r, err := regexp.Compile("^([a-zA-Z])(#)?([0-9]+)$")
	if err != nil {
		panic(err)
	}
	return r
}

func noteToValue(note string) (float64, bool) {
	n := noteRegexp.FindStringSubmatch(note)
	if n == nil {
		return 0, false
	}

	sharp := 0
	if n[2] == "#" {
		sharp = 1
	}

	octave, _ := strconv.Atoi(n[3])

	return float64(tones[n[1]]+sharp+(octave-4)*12) / 120, true
}

func jQuery(o interface{}) js.Object {
	return js.Global.Call("$", o)
}
