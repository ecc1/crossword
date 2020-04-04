package crossword

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/jung-kurt/gofpdf"
)

const (
	font            = "Helvetica"
	blackLevel      = 0.70 // ink-saving level: 1 = solid black squares
	marginPoints    = 18.0
	titlePoints     = 13.0
	minColumns      = 2
	maxColumns      = 6
	columnSepPoints = 3.0
	minCluePoints   = 6.0
	maxCluePoints   = 12.0
	cluePointsIncr  = 0.05
	interClueFrac   = 0.2
	lineWidthPoints = 0.5
	fullPageFrac    = 0.9
	minDownClues    = 2
	widthFrac       = 0.5
	heightFrac      = 0.6
)

type (
	RenderContext struct {
		Layouts    []Layout // in order of increasing NumColumns
		BestLayout int

		puz            *Puzzle
		pdf            *gofpdf.Fpdf
		pageWidth      float64 // page width
		pageHeight     float64 // page height
		margin         float64
		squareSize     float64 // grid square size
		renderWidth    float64 // rendered puzzle width
		renderHeight   float64 // rendered puzzle height
		puzzleLeft     float64
		puzzleTop      float64
		puzzleBottom   float64
		titleHeight    float64
		titleBottom    float64
		tall           bool
		wide           bool
		rendering      bool
		cluesFit       bool
		numColumns     int
		columnWidth    float64
		columnSep      float64
		cluePoints     float64
		clueLineHeight float64
		numberPoints   float64
		numberWidth    float64
		currentColumn  int // 1 .. numColumns
		leftMargin     float64
		x              float64
		y              float64
	}

	Layout struct {
		NumColumns int
		PointSize  float64
		Score      LayoutScore
	}

	Layouts []Layout
)

func (p *Puzzle) NewRenderContext(pdf *gofpdf.Fpdf) *RenderContext {
	r := RenderContext{puz: p, pdf: pdf}
	r.pageWidth, r.pageHeight = pdf.GetPageSize()
	r.margin = pdf.PointConvert(marginPoints)
	r.titleHeight = pdf.PointConvert(titlePoints)
	r.puzzleTop = r.margin + r.titleHeight + pdf.PointConvert(2)
	r.columnSep = pdf.PointConvert(columnSepPoints)
	r.setRenderSize()
	r.puzzleLeft = r.pageWidth - r.margin - r.renderWidth
	r.puzzleBottom = r.puzzleTop + r.renderHeight
	r.findLayouts()
	return &r
}

func (r *RenderContext) Render() {
	r.renderLayout(r.BestLayout)
}

func (r *RenderContext) RenderAll() {
	for i := range r.Layouts {
		r.renderLayout(i)
		r.markPage(i)
	}
}

func (r *RenderContext) setRenderSize() {
	puzWidth, puzHeight := float64(r.puz.Width), float64(r.puz.Height)
	puzAspect := puzWidth / puzHeight
	pageAspect := r.pageWidth / r.pageHeight
	if puzAspect <= pageAspect {
		// Scale puzzle width.
		r.renderWidth = widthFrac * (r.pageWidth - 2*r.margin)
		r.squareSize = r.renderWidth / puzWidth
		r.renderHeight = puzHeight * r.squareSize
		// Shrink puzzle height if it is too tall.
		maxHeight := r.pageHeight - (r.puzzleTop + r.margin)
		if r.renderHeight > maxHeight {
			r.renderHeight = maxHeight
			r.squareSize = r.renderHeight / puzHeight
			r.renderWidth = puzWidth * r.squareSize
		}
		r.tall = r.renderHeight > fullPageFrac*maxHeight
	} else {
		// Scale puzzle height.
		r.renderHeight = heightFrac * (r.pageHeight - (r.puzzleTop + r.margin))
		r.squareSize = r.renderHeight / puzHeight
		r.renderWidth = puzWidth * r.squareSize
		// Shrink puzzle width if it is too wide.
		maxWidth := r.pageWidth - 2*r.margin
		if r.renderWidth > maxWidth {
			r.renderWidth = maxWidth
			r.squareSize = r.renderWidth / puzWidth
			r.renderHeight = puzHeight * r.squareSize
		}
		r.wide = r.renderWidth > fullPageFrac*maxWidth
	}
}

func (r *RenderContext) renderLayout(i int) {
	r.pdf.AddPage()
	layout := r.Layouts[i]
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
	info := PuzzleString(puz.Author)
	pdf.SetFont(font, "", 0.9*titlePoints)
	w := pdf.GetStringWidth(info)
	x := r.pageWidth - r.margin - w
	y := r.margin + r.columnSep
	if r.rendering {
		pdf.Text(x, y, info)
	}
	title := PuzzleBytes(puz.Title)
	pdf.SetFont(font, "B", titlePoints)
	lines := pdf.SplitLines(title, x-2*r.columnSep)
	// Allow first line to extend up to author info.
	if r.rendering {
		pdf.Text(r.margin, y, string(lines[0]))
	}
	y += r.titleHeight
	lines = lines[1:]
	if len(lines) == 0 {
		r.titleBottom = math.Max(y+0.25*r.titleHeight, r.puzzleTop)
		return
	}
	if len(lines) > 1 && !r.wide {
		// Re-break subsequent lines to fit to the left of the grid.
		rest := bytes.Join(lines, []byte{' '})
		lines = pdf.SplitLines(rest, r.pageWidth-r.renderWidth-2*r.margin)
	}
	for _, v := range lines {
		if r.rendering {
			pdf.Text(r.margin, y, string(v))
		}
		y += r.titleHeight
	}
	r.titleBottom = math.Max(y+0.25*r.titleHeight, r.puzzleTop)
	if r.wide {
		r.puzzleTop = r.titleBottom
	}
}

func (r *RenderContext) drawGrid() {
	puz := r.puz
	puzWidth, puzHeight := float64(puz.Width), float64(puz.Height)
	pdf := r.pdf
	sq := r.squareSize
	// Transform coordinates so we can draw the puzzle grid with unit squares at (0,0).
	pdf.TransformBegin()
	pdf.TransformTranslate(r.puzzleLeft, r.puzzleTop)
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
			n := puz.SquareNumber(i, j)
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
	ch := r.clueLineHeight
	if r.leftMargin+r.columnWidth+r.columnSep <= r.puzzleLeft {
		r.y = r.titleBottom
	} else {
		r.y = r.puzzleBottom + r.columnSep + ch
	}
	r.x = r.leftMargin + r.numberWidth
	r.cluesFit = true
	puz := r.puz
	r.doClues(Across)
	// Make sure first few DOWN clues fit along with the heading.
	h := 0.0
	for _, n := range puz.Dir[Down].Numbers[:minDownClues] {
		c, _ := r.clueHeight(puz.Dir[Down].Clues[n])
		h += c
	}
	if r.y+1.75*ch+h > r.pageHeight-r.margin {
		r.nextColumn()
	} else {
		r.y += 0.5 * ch
	}
	r.doClues(Down)
}

// doClues renders the specified clues in the current layout.
// If r.rendering is false, the actual PDF rendering is not done,
// just the positioning, and r.cluesFit will indicate whether they fit on the page.
func (r *RenderContext) doClues(dir Direction) {
	if !r.cluesFit {
		return
	}
	d := &r.puz.Dir[dir]
	numbers := d.Numbers
	clues := d.Clues
	pdf := r.pdf
	ch := r.clueLineHeight
	rendering := r.rendering
	pdf.SetFont(font, "B", 0.8*r.cluePoints)
	if rendering {
		pdf.Text(r.x, r.y, dir.String())
	}
	r.y += 1.25 * ch
	pdf.SetFont(font, "", r.cluePoints)
	for _, n := range numbers {
		h, lines := r.clueHeight(clues[n])
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

// clueHeight calculates the height required to render a clue.
// It returns the height and the split lines for rendering.
func (r *RenderContext) clueHeight(clue string) (float64, [][]byte) {
	lines := r.pdf.SplitLines(PuzzleBytes(clue), r.columnWidth-r.numberWidth)
	h := (float64(len(lines)) + interClueFrac) * r.clueLineHeight
	return h, lines
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
		panic(fmt.Sprintf("clues do not fit %d-column format", r.numColumns))
	}
	r.leftMargin += r.columnWidth + r.columnSep
	if r.leftMargin+r.columnWidth+r.columnSep <= r.puzzleLeft {
		r.y = r.titleBottom
	} else {
		r.y = r.puzzleBottom + r.columnSep + r.clueLineHeight
	}
	r.x = r.leftMargin + r.numberWidth
}

func (r *RenderContext) findLayouts() {
	layouts := make(map[int]Layout)
	for n := minColumns; n <= maxColumns; n++ {
		for ps := maxCluePoints; ps >= minCluePoints; ps -= cluePointsIncr {
			r.setLayout(n, ps)
			r.rendering = false
			r.drawTitle()
			r.drawClues()
			r.rendering = true
			if !r.cluesFit {
				continue
			}
			score := r.layoutScore()
			cur := layouts[n].Score
			if !score.IsBetterThan(cur) {
				continue
			}
			layouts[n] = Layout{
				NumColumns: n,
				PointSize:  ps,
				Score:      score,
			}
		}
	}
	// Collect layouts and sort them.
	r.Layouts = make(Layouts, len(layouts))
	i := 0
	for _, layout := range layouts {
		r.Layouts[i] = layout
		i++
	}
	sort.Sort(Layouts(r.Layouts))
	// Find index of the best score.
	var bestScore LayoutScore
	bestLayout := -1
	for i, layout := range r.Layouts {
		if layout.Score.IsBetterThan(bestScore) {
			bestScore = layout.Score
			bestLayout = i
		}
	}
	r.BestLayout = bestLayout
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
	layout := r.Layouts[i]
	cw := r.getColumnWidth(layout.NumColumns)
	info := fmt.Sprintf("%.2fpt %.0f %s ", layout.PointSize, cw, layout.Score)
	pdf.SetFont(font, "", 7)
	x, y := r.pageWidth-r.margin, r.pageHeight-0.5*r.margin
	w := pdf.GetStringWidth(info)
	pdf.Text(x-w, y, info)
	if i == r.BestLayout {
		pdf.SetFont("ZapfDingbats", "", 14)
		pdf.Text(x, y, "\x34") // ✔
	}
}

// sort.Interface for Layouts using NumColumns as sort key.
func (v Layouts) Len() int           { return len(v) }
func (v Layouts) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v Layouts) Less(i, j int) bool { return v[i].NumColumns < v[j].NumColumns }
