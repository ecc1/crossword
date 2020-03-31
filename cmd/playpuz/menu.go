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

	showNotepadItem, _ := gtk.MenuItemNewWithLabel("Notepad")
	if puz.Notepad != "" {
		showNotepadItem.Connect("activate", showNotepad)
	} else {
		showNotepadItem.SetSensitive(false)
	}
	showNotepadItem.Show()
	menu.Append(showNotepadItem)

	checkMenu, _ := gtk.MenuNew()

	checkWordItem, _ := gtk.MenuItemNewWithLabel("Check word")
	checkWordItem.Connect("activate", checkWord)
	checkWordItem.Show()
	checkMenu.Append(checkWordItem)

	checkPuzzleItem, _ := gtk.MenuItemNewWithLabel("Check puzzle")
	checkPuzzleItem.Connect("activate", checkPuzzle)
	checkPuzzleItem.Show()
	checkMenu.Append(checkPuzzleItem)

	checkMenuItem, _ := gtk.MenuItemNewWithLabel("Check")
	checkMenuItem.SetSubmenu(checkMenu)
	checkMenuItem.Show()
	menu.Append(checkMenuItem)

	solveMenu, _ := gtk.MenuNew()

	solveWordItem, _ := gtk.MenuItemNewWithLabel("Solve word")
	solveWordItem.Connect("activate", solveWord)
	solveWordItem.Show()
	solveMenu.Append(solveWordItem)

	solvePuzzleItem, _ := gtk.MenuItemNewWithLabel("Solve puzzle")
	solvePuzzleItem.Connect("activate", solvePuzzle)
	solvePuzzleItem.Show()
	solveMenu.Append(solvePuzzleItem)

	solveMenuItem, _ := gtk.MenuItemNewWithLabel("Solve")
	solveMenuItem.SetSubmenu(solveMenu)
	solveMenuItem.Show()
	menu.Append(solveMenuItem)

	loadPuzzleItem, _ := gtk.MenuItemNewWithLabel("Load")
	loadPuzzleItem.Connect("activate", loadPuzzle)
	loadPuzzleItem.Show()
	menu.Append(loadPuzzleItem)

	savePuzzleItem, _ := gtk.MenuItemNewWithLabel("Save")
	savePuzzleItem.Connect("activate", savePuzzle)
	savePuzzleItem.Show()
	menu.Append(savePuzzleItem)

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
