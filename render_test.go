package acrosslite

import (
	"path"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

func TestRenderAllPuzzles(t *testing.T) {
	for _, file := range testFiles() {
		base := path.Base(file)
		t.Run(base, func(t *testing.T) {
			p, err := Read(file)
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			pdf := gofpdf.New("L", "pt", "Letter", "")
			rc := p.NewRenderContext(pdf)
			rc.Render()
			pdf.Close()
			if pdf.PageCount() != 1 {
				t.Errorf("Render(%s) produced %d pages", base, pdf.PageCount())
			}
		})
	}
}
