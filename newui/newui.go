package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/qml.v1"
	"gopkg.in/qml.v1/gl"
)

func main() {
	if err := qml.Run(nil, run); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var kinds = []string{
	"amp",
	"sin",
	"mul",
}

func run() error {
	qml.RegisterTypes("Sigourney", 1, 0, []qml.TypeSpec{{
		Init: func(r *Path, obj qml.Object) { r.Object = obj },
	}})

	engine := qml.NewEngine()

	base, err := engine.LoadFile("base.qml")
	if err != nil {
		return err
	}

	win := base.CreateWindow(nil)
	root := win.Root()

	ctx := engine.Context()
	ctx.SetVar("ctrl", &Controller{root: root})

	// Set up kinds.
	col := root.ObjectByName("kindColumn")
	for _, k := range kinds {
		kind := root.Object("kindComponent").Create(nil)
		kind.Set("kind", k)
		kind.Set("parent", col)
	}

	win.Show()
	win.Wait()

	return nil
}

type Controller struct {
	root qml.Object
}

func (c *Controller) OnDropKind(o qml.Object) {
	log.Println("drop", o.String("kind"), o.Int("x"), o.Int("y"))
}

type Path struct {
	qml.Object
	X1, Y1, X2, Y2 int
}

func (r *Path) Paint(p *qml.Painter) {
	height := r.Int("height")
	gl.LineWidth(2)
	gl.Color4f(0.0, 1.0, 0.0, 1.0)
	gl.Begin(gl.LINES)
	gl.Vertex2f(gl.Float(r.X1), gl.Float(height-r.Y1))
	gl.Vertex2f(gl.Float(r.X2), gl.Float(height-r.Y2))
	gl.End()
}
