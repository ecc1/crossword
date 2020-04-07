package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func buttonPress(x, y int, w gtk.IWidget, e *gdk.Event) {
	switch gdk.EventButtonNewFromEvent(e).Button() {
	case 1: // left
		if !puz.IsBlack(x, y) {
			setActive(x, y)
		}
	case 2: // middle
		menu.PopupAtPointer(e)
	case 3: // right
		if !puz.IsBlack(x, y) {
			changeDirection()
			setActive(x, y)
		}
	case 4: // scrollwheel up
	case 5: // scrollwheel down
	}
}

func keyPress(w gtk.IWidget, e *gdk.Event) bool {
	k := gdk.EventKeyNewFromEvent(e).KeyVal()
	if action, ok := keyAction[k]; ok {
		action()
	}
	// Indicate that the event has been consumed so that
	// the clue list boxes don't react to space, page down, etc.
	return true
}

var keyAction = map[uint]func(){
	' ':               eraseSquare,
	gdk.KEY_BackSpace: backspaceSquare,
	gdk.KEY_Delete:    backspaceSquare,
	gdk.KEY_Home:      moveHome,
	gdk.KEY_End:       moveEnd,
	gdk.KEY_Left:      moveLeft,
	gdk.KEY_Up:        moveUp,
	gdk.KEY_Right:     moveRight,
	gdk.KEY_Down:      moveDown,
}

func updateWith(c uint) func() {
	return func() {
		updateSquare(c)
		moveForward(true)
	}
}

func init() {
	// Add actions for letter keys.
	for k := uint('A'); k <= 'Z'; k++ {
		keyAction[k] = updateWith(k)
	}
	for k := uint('a'); k <= 'z'; k++ {
		keyAction[k] = updateWith('A' + (k - 'a'))
	}
}
