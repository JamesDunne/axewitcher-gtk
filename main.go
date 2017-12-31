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

	grid, _ := gtk.GridNew()
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	cbo, _ := gtk.ComboBoxTextNew()
	cbo.SetHExpand(true)
	grid.Add(cbo)

	stackMG, _ := gtk.StackNew()
	stackMG.SetHExpand(true)
	stackMG.SetVExpand(true)
	lblMG, _ := gtk.LabelNew("MG")
	stackMG.Add(lblMG)

	stackJD, _ := gtk.StackNew()
	stackJD.SetHExpand(true)
	stackJD.SetVExpand(true)
	lblJD, _ := gtk.LabelNew("JD")
	stackJD.Add(lblJD)

	gridSplit, _ := gtk.GridNew()
	gridSplit.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	gridSplit.SetHExpand(true)
	gridSplit.SetVExpand(true)

	gridSplit.Add(stackMG)
	gridSplit.Add(stackJD)
	grid.Add(gridSplit)
	win.Add(grid)

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

	// Add program names to combobox:
	for _, pr := range controller.Programs {
		log.Println(pr.Name)
		cbo.AppendText(pr.Name)
	}

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

		// Update UI elements:
		// TODO.

		// Redraw UI:
		//win.QueueDraw()
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

		// Update UI elements:
		// TODO.

		// Redraw UI:
		//win.QueueDraw()
	})

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	gtk.Main()
}
