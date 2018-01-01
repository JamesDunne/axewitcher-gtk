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
	controller *axewitcher.Controller
	amp        int
	s          *axewitcher.AmpState
	c          *axewitcher.AmpConfig

	frame       *gtk.Frame
	volume      *gtk.Scale
	btnDirty    *gtk.RadioButton
	btnClean    *gtk.RadioButton
	btnAcoustic *gtk.RadioButton
	FxButtons   [5]*gtk.ToggleButton
}

func (u *AmpUI) TopWidget() gtk.IWidget {
	return u.frame
}

func AmpUINew(controller *axewitcher.Controller, amp int, name string) *AmpUI {
	u := &AmpUI{
		controller: controller,
		amp:        amp,
		s:          &controller.Curr.Amp[amp],
		c:          &controller.Curr.AmpConfig[amp],
	}

	u.frame, _ = gtk.FrameNew(name)
	u.frame.SetLabelAlign(0.5, 0.5)
	u.frame.SetHExpand(true)
	u.frame.SetVExpand(true)

	grid, _ := gtk.GridNew()
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	// Add amp mode radio buttons:
	u.btnDirty, _ = gtk.RadioButtonNewWithLabel(nil, "dirty")
	u.btnDirty.SetMode(false)
	u.btnDirty.SetHExpand(true)
	u.btnDirty.Connect("toggled", func(btn *gtk.RadioButton) {
		if btn.GetActive() {
			u.s.Mode = axewitcher.AmpDirty
		}
	})
	u.btnClean, _ = gtk.RadioButtonNewWithLabelFromWidget(u.btnDirty, "clean")
	u.btnClean.SetMode(false)
	u.btnClean.SetHExpand(true)
	u.btnClean.Connect("toggled", func(btn *gtk.RadioButton) {
		if btn.GetActive() {
			u.s.Mode = axewitcher.AmpClean
		}
	})
	u.btnAcoustic, _ = gtk.RadioButtonNewWithLabelFromWidget(u.btnDirty, "acoustic")
	u.btnAcoustic.SetMode(false)
	u.btnAcoustic.SetHExpand(true)
	u.btnAcoustic.Connect("toggled", func(btn *gtk.RadioButton) {
		if btn.GetActive() {
			u.s.Mode = axewitcher.AmpAcoustic
		}
	})
	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	box.Add(u.btnDirty)
	box.Add(u.btnClean)
	box.Add(u.btnAcoustic)
	grid.Add(box)

	// Volume control:
	u.volume, _ = gtk.ScaleNewWithRange(
		gtk.ORIENTATION_HORIZONTAL,
		0,
		127,
		1)
	u.volume.SetHExpand(true)
	u.volume.SetValue(float64(u.s.Volume))
	u.volume.Connect("value-changed", func(r *gtk.Scale) {
		u.s.Volume = uint8(r.GetValue())
	})
	grid.Add(u.volume)

	box, _ = gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	// FX toggle buttons:
	for n := 0; n < 5; n++ {
		n := n
		u.FxButtons[n], _ = gtk.ToggleButtonNewWithLabel(u.c.Fx[n].Name)
		u.FxButtons[n].SetHExpand(true)
		u.FxButtons[n].Connect("toggled", func(btn *gtk.ToggleButton) {
			u.s.Fx[n].Enabled = btn.GetActive()
			u.controller.SendMidi()
		})
		box.Add(u.FxButtons[n])
	}
	grid.Add(box)

	u.frame.Add(grid)

	return u
}

func (u *AmpUI) Update() {
	// Update amp mode toggles:
	switch u.s.Mode {
	case axewitcher.AmpDirty:
		u.btnDirty.SetActive(true)
	case axewitcher.AmpClean:
		u.btnClean.SetActive(true)
	case axewitcher.AmpAcoustic:
		u.btnAcoustic.SetActive(true)
	}

	// Update volume range:
	u.volume.SetValue(float64(u.s.Volume))

	// Update Fx buttons:
	for a := 0; a < 5; a++ {
		u.FxButtons[a].SetActive(u.s.Fx[a].Enabled)
	}
}

type UI struct {
	win        *gtk.Window
	controller *axewitcher.Controller

	grid       *gtk.Grid
	cboProgram *gtk.ComboBoxText
	spinScene  *gtk.SpinButton
	ampUi      [2]*AmpUI
}

func NewUI(win *gtk.Window, controller *axewitcher.Controller) *UI {
	return &UI{
		win:        win,
		controller: controller,
	}
}

func (u *UI) Init() {
	// Create grid for UI:
	u.grid, _ = gtk.GridNew()
	u.grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	// Create combobox for program selection:
	u.cboProgram, _ = gtk.ComboBoxTextNew()
	u.cboProgram.SetHExpand(true)

	// Add program names to combobox:
	for _, pr := range u.controller.Programs {
		log.Println(pr.Name)
		u.cboProgram.AppendText(pr.Name)
	}

	u.cboProgram.SetActive(u.controller.Curr.PrIdx)
	u.cboProgram.Connect("changed", func(cbo *gtk.ComboBoxText) {
		if u.controller.Curr.PrIdx != cbo.GetActive() {
			u.controller.Curr.PrIdx = cbo.GetActive()
			u.controller.ActivateProgram()
			u.Update()
		}
	})

	spinAdjustment, _ := gtk.AdjustmentNew(
		float64(u.controller.Curr.SceneIdx),
		1, float64(len(u.controller.Curr.Pr.Scenes)),
		1, 1, 0)
	u.spinScene, _ = gtk.SpinButtonNew(spinAdjustment, 1.0, 0)
	u.spinScene.Connect("value-changed", func(spin *gtk.SpinButton) {
		if u.controller.Curr.SceneIdx != spin.GetValueAsInt()-1 {
			u.controller.Curr.SceneIdx = spin.GetValueAsInt() - 1
			u.controller.ActivateScene()
			u.Update()
		}
	})

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	box.Add(u.cboProgram)
	box.Add(u.spinScene)
	u.grid.Add(box)

	gridSplit, _ := gtk.GridNew()
	gridSplit.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	gridSplit.SetHExpand(true)
	gridSplit.SetVExpand(true)

	// Create UI widgets to represent amp states:
	u.ampUi = [2]*AmpUI{
		AmpUINew(u.controller, 0, "MG"),
		AmpUINew(u.controller, 1, "JD"),
	}

	gridSplit.Add(u.ampUi[0].TopWidget())
	gridSplit.Add(u.ampUi[1].TopWidget())

	u.grid.Add(gridSplit)
	u.win.Add(u.grid)
}

func (u *UI) Update() {
	// Update program selection:
	if u.cboProgram.GetActive() != u.controller.Curr.PrIdx {
		u.cboProgram.SetActive(u.controller.Curr.PrIdx)
		// Update scene spinner adjustment:
		u.spinScene.GetAdjustment().SetUpper(float64(len(u.controller.Curr.Pr.Scenes)))
	}

	// Update scene spinner:
	if u.spinScene.GetValueAsInt()-1 != u.controller.Curr.SceneIdx {
		u.spinScene.SetValue(float64(u.controller.Curr.SceneIdx + 1))
	}

	// Update UI elements:
	u.ampUi[0].Update()
	u.ampUi[1].Update()

	// Redraw UI:
	u.win.QueueDraw()
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

	// Create UI:
	ui := NewUI(win, controller)
	ui.Init()
	ui.Update()

	// Listen to key-press-events:
	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) bool {
		keyEvent := &gdk.EventKey{ev}

		var fswEvent axewitcher.FswEvent
		fswEvent.State = true
		fswEvent.Fsw = keyEventToFswEvent(keyEvent)
		if fswEvent.Fsw == axewitcher.FswNone {
			return false
		}

		// Handle the footswitch event with controller logic:
		controller.HandleFswEvent(fswEvent)

		ui.Update()
		return true
	})
	win.Connect("key-release-event", func(win *gtk.Window, ev *gdk.Event) bool {
		keyEvent := &gdk.EventKey{ev}

		var fswEvent axewitcher.FswEvent
		fswEvent.State = false
		fswEvent.Fsw = keyEventToFswEvent(keyEvent)
		if fswEvent.Fsw == axewitcher.FswNone {
			return false
		}

		// Handle the footswitch event with controller logic:
		controller.HandleFswEvent(fswEvent)

		ui.Update()
		return true
	})

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	gtk.Main()
}
