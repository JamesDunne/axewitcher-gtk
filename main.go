// maing
package main

import (
	"log"

	"github.com/JamesDunne/axewitcher"
	"github.com/gotk3/gotk3/gdk"
	//"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func keyEventToFswEvent(keyEvent *gdk.EventKey) axewitcher.FswButton {
	// PCsensor Footswitch3 defaults to A,B,C keys from left-to-right, mutually exclusive:
	switch keyEvent.KeyVal() {
	case gdk.KEY_A, gdk.KEY_a:
		// Reset:
		return axewitcher.FswReset
	case gdk.KEY_B, gdk.KEY_b:
		// Prev:
		return axewitcher.FswPrev
	case gdk.KEY_C, gdk.KEY_c:
		// Next:
		return axewitcher.FswNext
	default:
		// Ignore unknown button:
		return axewitcher.FswNone
	}
}

// Hold all UI widgets that should be updated from controller:
type AmpUI struct {
	s         *axewitcher.AmpState
	c         *axewitcher.AmpConfig
	topStack  *gtk.Stack
	Volume    *gtk.Scale
	FxButtons [5]*gtk.ToggleButton
}

func (u *AmpUI) TopWidget() gtk.IWidget {
	return u.topStack
}

func AmpUINew(name string, s *axewitcher.AmpState, c *axewitcher.AmpConfig) *AmpUI {
	u := &AmpUI{
		s: s,
		c: c,
	}

	u.topStack, _ = gtk.StackNew()
	u.topStack.SetHExpand(true)
	u.topStack.SetVExpand(true)

	grid, _ := gtk.GridNew()
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	lblName, _ := gtk.LabelNew(name)
	grid.Add(lblName)

	// Volume control:
	u.Volume, _ = gtk.ScaleNewWithRange(
		gtk.ORIENTATION_HORIZONTAL,
		0,
		127,
		1)
	u.Volume.SetHExpand(true)
	u.Volume.SetValue(float64(s.Volume))
	u.Volume.Connect("value-changed", func(r *gtk.Scale) {
		s.Volume = uint8(r.GetValue())
	})
	grid.Add(u.Volume)

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	// FX toggle buttons:
	for a := 0; a < 5; a++ {
		a := a
		u.FxButtons[a], _ = gtk.ToggleButtonNewWithLabel(c.Fx[a].Name)
		u.FxButtons[a].SetHExpand(true)
		u.FxButtons[a].Connect("toggled", func(btn *gtk.ToggleButton) {
			s.Fx[a].Enabled = btn.GetActive()
		})
		box.Add(u.FxButtons[a])
	}
	grid.Add(box)

	u.topStack.Add(grid)

	return u
}

func (u *AmpUI) Update() {
	// Update volume range:
	u.Volume.SetValue(float64(u.s.Volume))

	// Update Fx buttons:
	for a := 0; a < 5; a++ {
		u.FxButtons[a].SetActive(u.s.Fx[a].Enabled)
	}
}

func main() {
	gtk.Init(nil)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window: ", err)
	}
	win.SetTitle("Axewitcher 1.0")
	win.Connect("destroy", gtk.MainQuit)

	// Set the default window size for raspberry pi official display:
	win.SetDefaultSize(800, 480)

	// Create MIDI interface:
	midi, err := axewitcher.NewMidi()
	if err != nil {
		panic(err)
	}
	defer midi.Close()

	// Initialize controller:
	controller := axewitcher.NewController(midi)
	err = controller.Load()
	if err != nil {
		log.Fatal("Unable to load programs: ", err)
	}
	controller.Init()

	// Create grid for UI:
	grid, _ := gtk.GridNew()
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	// Create combobox for program selection:
	cboProgram, _ := gtk.ComboBoxTextNew()
	cboProgram.SetHExpand(true)

	// Add program names to combobox:
	for _, pr := range controller.Programs {
		log.Println(pr.Name)
		cboProgram.AppendText(pr.Name)
	}

	cboProgram.SetActive(controller.Curr.PrIdx)
	grid.Add(cboProgram)

	gridSplit, _ := gtk.GridNew()
	gridSplit.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	gridSplit.SetHExpand(true)
	gridSplit.SetVExpand(true)

	// Create UI widgets to represent amp states:
	ampUi := [2]*AmpUI{
		AmpUINew("MG", &controller.Curr.Amp[0], &controller.Curr.AmpConfig[0]),
		AmpUINew("JD", &controller.Curr.Amp[1], &controller.Curr.AmpConfig[1]),
	}

	gridSplit.Add(ampUi[0].TopWidget())
	gridSplit.Add(ampUi[1].TopWidget())

	grid.Add(gridSplit)
	win.Add(grid)

	updateUi := func() {
		if cboProgram.GetActive() != controller.Curr.PrIdx {
			cboProgram.SetActive(controller.Curr.PrIdx)
		}

		// Update UI elements:
		ampUi[0].Update()
		ampUi[1].Update()

		// Redraw UI:
		win.QueueDraw()
	}

	cboProgram.Connect("changed", func(cbo *gtk.ComboBoxText) {
		controller.Curr.PrIdx = cbo.GetActive()
		controller.ActivateProgram()
		updateUi()
	})

	// Listen to key-press-events:
	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{ev}

		var fswEvent axewitcher.FswEvent
		fswEvent.State = true
		fswEvent.Fsw = keyEventToFswEvent(keyEvent)
		if fswEvent.Fsw == axewitcher.FswNone {
			return
		}

		// Handle the footswitch event with controller logic:
		controller.HandleFswEvent(fswEvent)

		updateUi()
	})
	win.Connect("key-release-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{ev}

		var fswEvent axewitcher.FswEvent
		fswEvent.State = false
		fswEvent.Fsw = keyEventToFswEvent(keyEvent)
		if fswEvent.Fsw == axewitcher.FswNone {
			return
		}

		// Handle the footswitch event with controller logic:
		controller.HandleFswEvent(fswEvent)

		updateUi()
	})

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	gtk.Main()
}
