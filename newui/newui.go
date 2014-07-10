package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/qml.v1"
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
