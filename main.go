package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"image"
	"image/color"
	"log"

	"gioui.org/ui"
	"gioui.org/ui/app"
	"gioui.org/ui/f32"
	"gioui.org/ui/key"
	"gioui.org/ui/layout"
	"gioui.org/ui/measure"
	"gioui.org/ui/paint"
	"gioui.org/ui/text"
	"github.com/knsh14/astree"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/sfnt"
)

const initial = `
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, playground")
}`

func main() {
	go func() {
		w := app.NewWindow(app.WithTitle("AST Viewer"))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}

func loop(w *app.Window) error {
	var cfg app.Config
	mono, _ := sfnt.Parse(gomono.TTF)
	var faces measure.Faces
	ops := new(ui.Ops)
	edtr := &text.Editor{
		Face:       faces.For(mono, ui.Dp(22)),
		Submit:     false,
		SingleLine: false,
		Hint:       initial,
	}
	input := ""
	edtr.SetText(initial)
	l := &layout.List{Axis: layout.Vertical}
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case app.UpdateEvent:
			cfg = e.Config
			faces.Reset(&cfg)
			ops.Reset()
			cs := layout.RigidConstraints(e.Size)
			flex := layout.Flex{}
			flex.Init(ops, cs)

			// handle text update and update string to parse
			queue := w.Queue()
			for e, ok := edtr.Next(&cfg, queue); ok; e, ok = edtr.Next(&cfg, queue) {
				if _, ok = e.(text.ChangeEvent); ok {
					fs := token.NewFileSet()
					code := edtr.Text()
					f, err := parser.ParseFile(fs, "main.go", code, 0)
					if err != nil {
						input = err.Error()
					} else {
						var buf bytes.Buffer
						astree.File(&buf, fs, f)
						input = buf.String()
					}
				}
			}

			// layout editor view
			cs = flex.Flexible(0.49)
			dims := edtr.Layout(&cfg, queue, ops, cs)
			leftside := flex.End(dims)

			// draw a black line to separate
			cs = flex.Flexible(0.01)
			square := f32.Rectangle{
				Max: f32.Point{
					X: float32(cs.Width.Max),
					Y: float32(cs.Height.Max),
				},
			}
			paint.ColorOp{Color: color.RGBA{A: 0xff}}.Add(ops)
			paint.PaintOp{Rect: square}.Add(ops)
			dims = layout.Dimensions{Size: image.Point{X: cs.Width.Max, Y: cs.Height.Max}}
			centerline := flex.End(dims)

			// layout tree to right side
			cs = flex.Flexible(0.5)
			if l.Dragging() {
				key.HideInputOp{}.Add(ops)
			}
			for l.Init(&cfg, queue, ops, cs, 1); l.More(); l.Next() {
				dims = text.Label{
					Face: faces.For(mono, ui.Dp(16)),
					Text: input,
				}.Layout(ops, l.Constraints())
				l.End(dims)
			}
			rightside := flex.End(l.Layout())

			dims = flex.Layout(leftside, centerline, rightside)
			w.Update(ops)
		}
	}
}
