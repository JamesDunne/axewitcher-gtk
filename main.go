// maing
package main

import (
	"log"

	"github.com/JamesDunne/axewitcher"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
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

	// Listen to key-press-events:
	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{ev}
		// PCsensor Footswitch3 defaults to A,B,C keys from left-to-right, mutually exclusive:
		var fswEvent axewitcher.FswEvent
		switch keyEvent.KeyVal() {
		case gdk.KEY_A:
			// Reset:
			fswEvent.Fsw = axewitcher.FswReset
			break
		case gdk.KEY_B:
			// Prev:
			fswEvent.Fsw = axewitcher.FswPrev
			break
		case gdk.KEY_C:
			// Next:
			fswEvent.Fsw = axewitcher.FswNext
			break
		default:
			// Ignore unknown button:
			return
		}

		// Handle the footswitch event with controller logic:
		controller.HandleFswEvent(fswEvent)

		// Update UI elements:
		// TODO.

		// Redraw UI:
		win.QueueDraw()
	})

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	gtk.Main()
}
