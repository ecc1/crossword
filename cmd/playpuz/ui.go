package main

import (
	"fmt"
	"math"

	"github.com/ecc1/crossword"
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const (
	textFont = "Sans"

	// Relative to a unit square.
	lineWidth     = 0.015
	innerSep      = 0.075
	largeFontSize = 0.500
	smallFontSize = 0.300

	minCellSize = 75
	clueWidth   = 400
	gridWidth   = 1200
)

var (
	puzWidth  float64
	puzHeight float64

	window    *gtk.Window
	grid      *gtk.Grid
	clueLists = make([]*gtk.ListBox, 2)

	maxWidth  int
	maxHeight int

	// RGBA values for the backgrounds of puzzle squares.
	blackColor  = []float64{0, 0, 0, 1}
	normalColor = []float64{1, 1, 1, 1}
	activeColor = []float64{0, 1, 0.5, 1}
	wordColor   = []float64{0.75, 0.75, 0.75, 1}
	wrongColor  = []float64{0.9, 0.3, 0.3, 1}
)

func initUI() {
	puzWidth = float64(puz.Width)
	puzHeight = float64(puz.Height)
	gtk.Init(nil)
	window, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	window.SetTitle(puz.Title)
	setGeometry()
	window.Connect("destroy", gtk.MainQuit)
	window.Connect("key-press-event", keyPress)
	window.Add(makeTopLevel())
	makeMenu()
	window.ShowAll()
	setActivePos(cur)
}

func setGeometry() {
	d, _ := window.GetScreen().GetDisplay()
	m, _ := d.GetPrimaryMonitor()
	r := m.GetGeometry()
	maxWidth = 3 * r.GetWidth() / 4
	maxHeight = 3 * r.GetHeight() / 4
	window.SetDefaultSize(-1, maxHeight/puz.Height)
	window.SetPosition(gtk.WIN_POS_MOUSE)
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func makeTopLevel() gtk.IWidget {
	p, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	p.SetWideHandle(true)
	p.Pack1(makeClues(Across), true, false)
	q, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	q.SetWideHandle(true)
	q.Pack1(makeGrid(), true, false)
	q.Pack2(makeClues(Down), true, false)
	p.Pack2(q, true, false)
	return p
}

func makeGrid() gtk.IWidget {
	grid, _ = gtk.GridNew()
	grid.SetRowHomogeneous(true)
	grid.SetRowSpacing(0)
	grid.SetColumnHomogeneous(true)
	grid.SetColumnSpacing(0)
	for y := 0; y < puz.Height; y++ {
		for x := 0; x < puz.Width; x++ {
			attachCell(x, y)
		}
	}
	r := float32(puzWidth / puzHeight)
	a, _ := gtk.AspectFrameNew("", 0.5, 0.5, r, false)
	a.Add(grid)
	w := min(puz.Width*minCellSize, maxWidth)
	h := min(puz.Height*minCellSize, maxHeight)
	a.SetSizeRequest(w, h)
	return a
}

func makeClues(dir crossword.Direction) gtk.IWidget {
	d := &puz.Dir[dir]
	h, _ := gtk.LabelNew("")
	h.SetMarkup(fmt.Sprintf("<b>%s</b>", dir))
	h.SetWidthChars(10)
	b, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	b.PackStart(h, false, false, 0)
	clueList, _ := gtk.ListBoxNew()
	clueList.SetActivateOnSingleClick(true)
	for _, n := range d.Numbers {
		clueList.Add(makeClue(dir, n))
	}
	s, _ := gtk.ScrolledWindowNew(nil, nil)
	s.SetOverlayScrolling(false)
	s.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	s.Add(clueList)
	// Make sure that the clue will scroll into view when selected.
	clueList.SetFocusVAdjustment(s.GetVAdjustment())
	clueList.Connect("row-activated", func(w gtk.IWidget, row *gtk.ListBoxRow) { chooseRow(dir, w, row) })
	b.PackStart(s, true, true, 0)
	clueLists[dir] = clueList
	b.SetSizeRequest(clueWidth, -1)
	return b
}

func makeClue(dir crossword.Direction, n int) gtk.IWidget {
	cl, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	num, _ := gtk.LabelNew("")
	num.SetMarkup(fmt.Sprintf("<small><b>%d</b></small>  ", n))
	// Make this wide enough for 3-digit numbers plus space with the above markup.
	num.SetWidthChars(4)
	num.SetXAlign(1)
	num.SetYAlign(0)
	cl.PackStart(num, false, false, 0)
	clueStr := puz.Dir[dir].Clues[n]
	lbl, _ := gtk.LabelNew(clueStr)
	lbl.SetXAlign(0)
	lbl.SetLineWrap(true)
	lbl.SetJustify(gtk.JUSTIFY_LEFT)
	cl.PackStart(lbl, false, false, 0)
	return cl
}

func chooseRow(dir crossword.Direction, w gtk.IWidget, row *gtk.ListBoxRow) {
	i := row.GetIndex()
	n := puz.Dir[dir].Numbers[i]
	activateFromClue(dir, n)
}

func selectClue(dir crossword.Direction, i int) {
	lb := clueLists[dir]
	cl := lb.GetRowAtIndex(i)
	lb.SelectRow(cl)
	cl.GrabFocus()
}

func attachCell(x int, y int) {
	d, _ := gtk.DrawingAreaNew()
	d.SetHExpand(true)
	d.SetVExpand(true)
	d.Connect("draw", func(d *gtk.DrawingArea, c *cairo.Context) { drawCell(x, y, d, c) })
	eb, _ := gtk.EventBoxNew()
	eb.Add(d)
	eb.Connect("button-press-event", func(w gtk.IWidget, e *gdk.Event) { buttonPress(x, y, w, e) })
	grid.Attach(eb, x+1, y+1, 1, 1)
}

func drawCell(x int, y int, d *gtk.DrawingArea, c *cairo.Context) {
	// Transform cell to unit square.
	// Don't assume allocated width is exactly equal to allocated height;
	// AspectFrame will keep them close enough but not necessarily equal.
	c.Scale(float64(d.GetAllocatedWidth()), float64(d.GetAllocatedHeight()))
	c.SetLineWidth(0)
	if puz.IsBlack(x, y) {
		setColor(c, blackColor)
		c.Rectangle(0, 0, 1, 1)
		c.Fill()
		return
	}
	// Background color.
	bg := normalColor
	if cells[y][x] == wrongSquare {
		bg = wrongColor
	} else if isActive(x, y) {
		bg = activeColor
	} else if inActiveWord(x, y) {
		bg = wordColor
	}
	setColor(c, bg)
	c.Rectangle(0, 0, 1, 1)
	c.Fill()
	// Grid lines.
	c.SetLineWidth(lineWidth)
	setColor(c, blackColor)
	c.MoveTo(0, 1)
	c.LineTo(1, 1)
	c.LineTo(1, 0)
	if x == 0 {
		c.MoveTo(0, 0)
		c.LineTo(0, 1)
	}
	if y == 0 {
		c.MoveTo(0, 0)
		c.LineTo(1, 0)
	}
	c.Stroke()
	// Cell number.
	n := puz.SquareNumber(x, y)
	if n != 0 {
		c.SelectFontFace(textFont, cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
		c.SetFontSize(smallFontSize)
		num := fmt.Sprintf("%d", n)
		t := c.TextExtents(num)
		c.MoveTo(innerSep, innerSep+t.Height)
		c.ShowText(num)
	}
	// Circle, if any.
	if puz.IsCircled(x, y) {
		c.NewPath()
		Δ := lineWidth / 2
		c.Arc(0.5+Δ, 0.5+Δ, 0.5-2*Δ, 0, 2*math.Pi)
		c.Stroke()
	}
	// Cell contents.
	c.SelectFontFace(textFont, cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	c.SetFontSize(largeFontSize)
	s := fmt.Sprintf("%c", cells[y][x])
	t := c.TextExtents(s)
	// Ignore t.Height so character baselines are aligned.
	c.MoveTo(0.5-t.Width/2, 0.75)
	c.ShowText(s)
}

func setColor(c *cairo.Context, color []float64) {
	c.SetSourceRGBA(color[0], color[1], color[2], color[3])
}

func redrawCell(x, y int) {
	w, _ := grid.GetChildAt(x+1, y+1)
	w.QueueDraw()
}

func runUI() {
	gtk.Main()
}

func winnerWinner() {
	dialog := gtk.MessageDialogNewWithMarkup(window, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "")
	dialog.SetMarkup("<b>You have solved the puzzle.</b>")
	dialog.Run()
	dialog.Destroy()
}
