package crossword

import (
	"path"
	"sync"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

func TestRenderSinglePagePuzzles(t *testing.T) {
	renderPuzzles(t, false)
}

func TestRenderMultiPagePuzzles(t *testing.T) {
	renderPuzzles(t, true)
}

func renderPuzzles(t *testing.T, multi bool) {
	for _, base := range testFiles() {
		t.Run(base, func(t *testing.T) {
			file := path.Join(testDataDir, base)
			p, err := Read(file)
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			pdf := gofpdf.New("L", "pt", "Letter", "")
			rc := p.NewRenderContext(pdf)
			var pagesWanted int
			if multi {
				rc.RenderAll()
				pagesWanted = len(rc.Layouts)
			} else {
				rc.Render()
				pagesWanted = 1
			}
			n := pdf.PageCount()
			pdf.Close()
			if n != pagesWanted {
				t.Errorf("rendering %q produced %d pages, want %d", base, n, pagesWanted)
			}
		})
	}
}

func TestAllLayouts(t *testing.T) {
	testLayouts(t, testFiles(), "/tmp/crossword_render_test.pdf")
}

func TestEccentricLayouts(t *testing.T) {
	testLayouts(t, eccentricPuzzles, "/tmp/eccentric_render_test.pdf")
}

func testLayouts(t *testing.T, files []string, outputFile string) {
	pdf := gofpdf.New("L", "pt", "Letter", "")
	var m sync.Mutex
	for _, base := range files {
		t.Run(base, func(t *testing.T) {
			best := bestLayout[base]
			if best == 0 {
				t.Errorf("file %q not listed in layout cases", base)
				return
			}
			file := path.Join(testDataDir, base)
			p, err := Read(file)
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			m.Lock()
			rc := p.NewRenderContext(pdf)
			rc.RenderAll()
			m.Unlock()
			n := rc.Layouts[rc.BestLayout].NumColumns
			if n != best {
				t.Errorf("got best layout = %d columns, want %d columns", n, best)
			}
		})
	}
	// Save PDF for further examination.
	pdf.OutputFileAndClose(outputFile)
}

const benchmarkPuzzle = "Aug0810.puz"

func BenchmarkRender(b *testing.B) {
	p, err := Read(path.Join(testDataDir, benchmarkPuzzle))
	if err != nil {
		b.Errorf("%s", err)
		return
	}
	pdf := gofpdf.New("L", "pt", "Letter", "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.NewRenderContext(pdf)
	}
}
