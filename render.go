package acrosslite

import (
	"fmt"
	"math"
	"os"

	"github.com/jung-kurt/gofpdf"
)

const (
	font            = "Helvetica"
	marginPoints    = 18.0
	titlePoints     = 13.0
	minColumns      = 2
	maxColumns      = 6
	columnSepPoints = 3.0
	minCluePoints   = 6.0
	maxCluePoints   = 12.0
	cluePointsIncr  = 0.1
	interClueFrac   = 0.2
	lineWidthPoints = 0.5
	blackLevel      = 0.70

	debug = true
)

type (
	RenderContext struct {
		puz            *Puzzle
		pdf            *gofpdf.Fpdf
		layouts        []Layout // in order of increasing NumColumns
		bestLayout     int
		pageWidth      float64 // page width
		pageHeight     float64 // page height
		squareSize     float64 // grid square size
		renderWidth    float64 // rendered puzzle width
		renderHeight   float64 // rendered puzzle height
		margin         float64
		top            float64
		titleHeight    float64
		titleY         float64
		tall           bool
		rendering      bool
		cluesFit       bool
		numColumns     int
		columnWidth    float64
		columnSep      float64
		cluePoints     float64
		clueLineHeight float64
		numberPoints   float64
		numberWidth    float64
		currentColumn  int
		leftMargin     float64
		x              float64
		y              float64
	}

	Layout struct {
		NumColumns int
		PointSize  float64
		Score      float64
	}
)

func (p *Puzzle) NewRenderContext(pdf *gofpdf.Fpdf) *RenderContext {
	r := RenderContext{puz: p, pdf: pdf}
	r.pageWidth, r.pageHeight = pdf.GetPageSize()
	r.margin = pdf.PointConvert(marginPoints)
	r.titleHeight = pdf.PointConvert(titlePoints)
	r.top = r.margin + r.titleHeight + pdf.PointConvert(2)
	// Scale puzzle to half the page width.
	puzWidth, puzHeight := float64(p.Width), float64(p.Height)
	r.renderWidth = r.pageWidth/2 - r.margin
	r.squareSize = r.renderWidth / puzWidth
	r.renderHeight = puzHeight * r.squareSize
	r.columnSep = pdf.PointConvert(columnSepPoints)
	// Shrink puzzle to page height if it is too tall.
	if r.renderHeight+2*r.margin > r.pageHeight {
		r.tall = true
		r.renderHeight = r.pageHeight - (r.top + r.margin)
		r.squareSize = r.renderHeight / puzHeight
		r.renderWidth = puzWidth * r.squareSize
	}
	r.findLayouts()
	return &r
}

func (r *RenderContext) Render() {
	r.renderLayout(r.bestLayout)
}

func (r *RenderContext) RenderAll() {
	for i := range r.layouts {
		r.renderLayout(i)
		r.markPage(i)
	}
}

func (r *RenderContext) Layouts() []Layout {
	return r.layouts
}

func (r *RenderContext) renderLayout(i int) {
	r.pdf.AddPage()
	layout := r.layouts[i]
	r.setLayout(layout.NumColumns, layout.PointSize)
	r.drawTitle()
	r.drawGrid()
	r.drawClues()
}

// drawTitle draws title and author information at the top of the page and
// records the position of the bottom of the (possibly multi-line) title in r.titleY.
func (r *RenderContext) drawTitle() {
	puz := r.puz
	pdf := r.pdf
	rendering := r.rendering
	m := r.margin
	y0 := m + r.columnSep
	title := PuzzleBytes(puz.Title)
	pdf.SetFont(font, "B", titlePoints)
	lines := pdf.SplitLines(title, r.pageWidth-(r.renderWidth+m)-m)
	y := y0
	for _, v := range lines {
		if rendering {
			pdf.Text(m, y, string(v))
		}
		y += r.titleHeight
	}
	r.titleY = math.Max(y+0.25*r.titleHeight, r.top)
	if !rendering {
		return
	}
	info := string(PuzzleBytes(puz.Author))
	pdf.SetFont(font, "", 0.9*titlePoints)
	w := pdf.GetStringWidth(info)
	pdf.Text(r.pageWidth-m-w, y0, info)
}

func (r *RenderContext) drawGrid() {
	puz := r.puz
	puzWidth, puzHeight := float64(puz.Width), float64(puz.Height)
	pdf := r.pdf
	sq := r.squareSize
	// Transform coordinates so we can draw the puzzle grid with unit squares at (0,0).
	pdf.TransformBegin()
	pdf.TransformTranslate(r.pageWidth-(r.renderWidth+r.margin), r.top)
	pdf.TransformScale(100*sq, 100*sq, 0, 0) // scale factors are percentages
	pdf.SetLineWidth(pdf.PointConvert(lineWidthPoints) / sq)
	// Draw horizontal grid lines.
	for y := 0.0; y <= puzHeight; y++ {
		pdf.Line(0, y, puzWidth, y)
	}
	// Draw vertical grid lines.
	for x := 0.0; x <= puzWidth; x++ {
		pdf.Line(x, 0, x, puzHeight)
	}
	// Draw black squares, numbers, and circles.
	v := int(math.Round((1 - blackLevel) * 255))
	pdf.SetFillColor(v, v, v)
	numberSize := 0.3 // scaled by 1/sq
	pdf.SetFont(font, "", numberSize)
	for y := 0.0; y < puzHeight; y++ {
		for x := 0.0; x < puzWidth; x++ {
			i, j := int(x), int(y)
			if puz.IsBlack(i, j) {
				pdf.Rect(x, y, 1, 1, "F")
				continue
			}
			n := puz.CellNumber(i, j)
			if n != 0 {
				pdf.Text(x+0.05, y+numberSize, fmt.Sprintf("%d", n))
			}
			if puz.IsCircled(i, j) {
				lw := pdf.GetLineWidth()
				pdf.SetLineWidth(lw / 2)
				pdf.Circle(x+0.5, y+0.5, 0.5, "D")
				pdf.SetLineWidth(lw)
			}
		}
	}
	pdf.TransformEnd()
}

// drawClues renders the Across and Down clues in the current layout.
// If r.rendering is false, the actual PDF rendering is not done,
// just the positioning, and upon return the r.cluesFit field
// indicates whether they fit on the page.
func (r *RenderContext) drawClues() {
	r.currentColumn = 1
	r.leftMargin = r.margin
	r.x = r.leftMargin + r.numberWidth
	r.y = r.titleY
	r.cluesFit = true
	puz := r.puz
	ch := r.clueLineHeight
	r.doClues("ACROSS", puz.AcrossNumbers, puz.AcrossClues)
	if r.y+2*ch > r.pageHeight-r.margin {
		r.nextColumn()
	} else {
		r.y += 0.5 * ch
	}
	r.doClues("DOWN", puz.DownNumbers, puz.DownClues)
}

// doClues renders the specified clues in the current layout.
// If r.rendering is false, the actual PDF rendering is not done,
// just the positioning, and r.cluesFit will indicate whether they fit on the page.
func (r *RenderContext) doClues(heading string, numbers []int, clues IndexedStrings) {
	if !r.cluesFit {
		return
	}
	pdf := r.pdf
	colWidth := r.columnWidth
	ch := r.clueLineHeight
	numWidth := r.numberWidth
	rendering := r.rendering
	pdf.SetFont(font, "B", r.cluePoints)
	if rendering {
		pdf.Text(r.x, r.y, heading)
	}
	r.y += 1.25 * ch
	for _, n := range numbers {
		clue := PuzzleBytes(clues[n])
		lines := pdf.SplitLines(clue, colWidth-numWidth)
		h := float64(len(lines))*ch + interClueFrac*ch
		if r.y+h > r.pageHeight-r.margin {
			r.nextColumn()
			if !r.cluesFit {
				return
			}
		}
		s := makeClueNumber(n)
		pdf.SetFont(font, "B", r.numberPoints)
		w := pdf.GetStringWidth(s)
		if rendering {
			pdf.Text(r.x-w, r.y, s)
		}
		pdf.SetFont(font, "", r.cluePoints)
		for _, v := range lines {
			if rendering {
				pdf.Text(r.x, r.y, string(v))
			}
			r.y += ch
		}
		r.y += interClueFrac * ch
	}
}

func (r *RenderContext) setNumberWidth() {
	pdf := r.pdf
	r.numberPoints = 0.8 * r.cluePoints
	pdf.SetFont(font, "B", r.numberPoints)
	max := 0.0
	for n := 1; n <= r.puz.Height*r.puz.Width; n++ {
		w := pdf.GetStringWidth(makeClueNumber(n))
		if max < w {
			max = w
		}
	}
	r.numberWidth = max
}

func makeClueNumber(n int) string {
	return fmt.Sprintf("%d  ", n)
}

func (r *RenderContext) nextColumn() {
	r.currentColumn++
	if r.currentColumn > r.numColumns {
		if !r.rendering {
			r.cluesFit = false
			return
		}
		// Bug: this layout should have been avoided by findLayouts().
		msg := fmt.Sprintf("clues do not fit %d-column format", r.numColumns)
		if !debug {
			panic(msg)
		}
		fmt.Fprintf(os.Stderr, "%s [%s]\n", msg, r.puz.Title)
		// Ensure that tests for single-page rendering will fail.
		r.pdf.AddPage()
		r.leftMargin = r.margin
		r.y = r.margin
		return
	}
	r.leftMargin += r.columnWidth + r.columnSep
	r.x = r.leftMargin + r.numberWidth
	if !r.tall && r.currentColumn > r.numColumns/2 {
		r.y = r.top + r.renderHeight + r.columnSep + r.clueLineHeight
	} else {
		r.y = r.titleY
	}
}

func (r *RenderContext) findLayouts() {
	colIncr := 1
	if !r.tall {
		// Don't use odd numbers of columns for non-tall puzzles.
		colIncr = 2
	}
	for n := minColumns; n <= maxColumns; n += colIncr {
		for ps := maxCluePoints; ps >= minCluePoints; ps -= cluePointsIncr {
			r.setLayout(n, ps)
			r.rendering = false
			r.drawTitle()
			r.drawClues()
			r.rendering = true
			if r.cluesFit {
				r.layouts = append(r.layouts, Layout{
					NumColumns: n,
					PointSize:  ps,
					Score:      r.layoutScore(),
				})
				break
			}
		}
	}
	// Find best layout.
	bestScore := 0.0
	bestLayout := -1
	for i, layout := range r.layouts {
		if layout.Score > bestScore {
			bestScore = layout.Score
			bestLayout = i
		}
	}
	r.bestLayout = bestLayout
}

func (r *RenderContext) setLayout(numCols int, cluePoints float64) {
	r.numColumns = numCols
	r.columnWidth = r.getColumnWidth(numCols)
	r.cluePoints = cluePoints
	r.clueLineHeight = r.pdf.PointConvert(cluePoints)
	r.setNumberWidth()
}

func (r *RenderContext) getColumnWidth(n int) float64 {
	w := r.pageWidth - 2*r.margin - float64(n+1)*r.columnSep
	if r.tall {
		w -= r.renderWidth
	}
	return w / float64(n)
}

func (r *RenderContext) markPage(i int) {
	pdf := r.pdf
	layout := r.layouts[i]
	info := fmt.Sprintf("%0.2fpt %0.3f ", layout.PointSize, layout.Score)
	pdf.SetFont(font, "", 7)
	w := pdf.GetStringWidth(info)
	x, y := r.pageWidth-r.margin, r.pageHeight-0.5*r.margin
	pdf.Text(x-w, y, info)
	if i == r.bestLayout {
		pdf.SetFont("Symbol", "", 13)
		pdf.Text(x, y, "\x34")
	}
}

// layoutScore calculates a "goodness" score for the current layout.
// Larger point size is preferred, then fewer columns.
// TO DO: add penalty for blank or under-filled columns.
func (r *RenderContext) layoutScore() float64 {
	n := float64(maxColumns-r.numColumns) / float64(maxColumns-minColumns)
	p := (r.cluePoints - minCluePoints) / (maxCluePoints - minCluePoints)
	return 10*p + n
}
