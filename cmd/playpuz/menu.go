package main

import (
	"io/ioutil"

	"github.com/gotk3/gotk3/gtk"
)

var (
	menu *gtk.Menu
)

func makeMenu() {
	menu, _ = gtk.MenuNew()

	notepad, _ := gtk.MenuItemNewWithLabel("Notepad")
	notepad.Connect("activate", showNotepad)
	notepad.Show()
	menu.Append(notepad)

	load, _ := gtk.MenuItemNewWithLabel("Load")
	load.Connect("activate", loadPuzzle)
	load.Show()
	menu.Append(load)

	save, _ := gtk.MenuItemNewWithLabel("Save")
	save.Connect("activate", savePuzzle)
	save.Show()
	menu.Append(save)

	quit, _ := gtk.MenuItemNewWithLabel("Quit")
	quit.Connect("activate", gtk.MainQuit)
	quit.Show()
	menu.Append(quit)
}

func showNotepad() {
	text := puz.Notepad
	if text == "" {
		text = "(This puzzle has no notepad.)"
	}
	dialog := gtk.MessageDialogNew(window, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "%s", text)
	dialog.Run()
	dialog.Destroy()
}

func loadPuzzle() {
	dialog, _ := gtk.FileChooserDialogNewWith2Buttons("Load Puzzle", window, gtk.FILE_CHOOSER_ACTION_OPEN,
		"Cancel", gtk.RESPONSE_CANCEL,
		"Load", gtk.RESPONSE_ACCEPT)
	res := dialog.Run()
	if res != gtk.RESPONSE_ACCEPT {
		dialog.Destroy()
		return
	}
	filename := dialog.GetFilename()
	dialog.Destroy()
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		popupError(err)
		return
	}
	err = setContents(contents)
	if err != nil {
		popupError(err)
		return
	}
	if puzzleIsSolved() {
		winnerWinner()
	}
}

func savePuzzle() {
	dialog, _ := gtk.FileChooserDialogNewWith2Buttons("Save Puzzle", window, gtk.FILE_CHOOSER_ACTION_SAVE,
		"Cancel", gtk.RESPONSE_CANCEL,
		"Save", gtk.RESPONSE_ACCEPT)
	dialog.SetDoOverwriteConfirmation(true)
	res := dialog.Run()
	if res != gtk.RESPONSE_ACCEPT {
		dialog.Destroy()
		return
	}
	filename := dialog.GetFilename()
	dialog.Destroy()
	err := ioutil.WriteFile(filename, getContents(), 0644)
	if err != nil {
		popupError(err)
	}
}

func popupError(err error) {
	dialog := gtk.MessageDialogNew(window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "Error: %s", err)
	dialog.Run()
	dialog.Destroy()
}
