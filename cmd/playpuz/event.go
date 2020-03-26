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

func keyPress(w gtk.IWidget, e *gdk.Event) {
	k := gdk.EventKeyNewFromEvent(e).KeyVal()
	if 'A' <= k && k <= 'Z' {
		updateCell(k)
		moveForward(true)
		return
	}
	if 'a' <= k && k <= 'z' {
		updateCell('A' + (k - 'a'))
		moveForward(true)
		return
	}
	switch k {
	case ' ':
		updateCell(emptySquare)
	case gdk.KEY_BackSpace, gdk.KEY_Delete:
		updateCell(emptySquare)
		moveBackward(false)
	case gdk.KEY_Home:
		moveHome()
	case gdk.KEY_Left:
		moveLeft()
	case gdk.KEY_Up:
		moveUp()
	case gdk.KEY_Right:
		moveRight()
	case gdk.KEY_Down:
		moveDown()
	case gdk.KEY_End:
		moveEnd()
	}
}
