package acrosslite

import (
	"path"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

func TestRenderSinglePagePuzzles(t *testing.T) {
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
			n := pdf.PageCount()
			if n != 1 {
				t.Errorf("Render(%s) produced %d pages, want 1", base, n)
			}
		})
	}
}

func TestRenderMultiPagePuzzles(t *testing.T) {
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
			rc.RenderAll()
			pdf.Close()
			n := pdf.PageCount()
			if n != len(rc.Layouts) {
				t.Errorf("RenderAll(%s) produced %d pages, want %d", base, n, len(rc.Layouts))
			}
		})
	}
}

func TestLayouts(t *testing.T) {
	cases := []struct {
		file string
		best int // number of columns
	}{
		{"Apr0310.puz", 4},
		{"Apr0410.puz", 6},
		{"Apr1010.puz", 4},
		{"Apr1110.puz", 6},
		{"Apr1710.puz", 4},
		{"Apr1810.puz", 6},
		{"Apr2410.puz", 4},
		{"Apr2510.puz", 6},
		{"Aug0110.puz", 6},
		{"Aug0710.puz", 4},
		{"Aug0810.puz", 6},
		{"Aug1410.puz", 4},
		{"Aug1510.puz", 6},
		{"Aug2110.puz", 4},
		{"Aug2210.puz", 6},
		{"Aug2810.puz", 4},
		{"Aug2910.puz", 6},
		{"Dec0410.puz", 4},
		{"Dec0510.puz", 6},
		{"Dec1110.puz", 4},
		{"Dec1210.puz", 6},
		{"Dec1810.puz", 4},
		{"Dec1910.puz", 6},
		{"Dec2510.puz", 4},
		{"Dec2610.puz", 6},
		{"Feb0308.puz", 6},
		{"Feb0610.puz", 4},
		{"Feb0710.puz", 6},
		{"Feb1310.puz", 4},
		{"Feb1410.puz", 6},
		{"Feb2010.puz", 4},
		{"Feb2110.puz", 6},
		{"Feb2710.puz", 6},
		{"Feb2810.puz", 6},
		{"Jan0210.puz", 4},
		{"Jan0310.puz", 6},
		{"Jan0910.puz", 4},
		{"Jan1010.puz", 6},
		{"Jan1610.puz", 4},
		{"Jan1710.puz", 6},
		{"Jan2310.puz", 4},
		{"Jan2410.puz", 6},
		{"Jan3010.puz", 4},
		{"Jan3110.puz", 6},
		{"Jul0310.puz", 4},
		{"Jul0410.puz", 4},
		{"Jul1010.puz", 4},
		{"Jul1110.puz", 6},
		{"Jul1710.puz", 4},
		{"Jul1810.puz", 6},
		{"Jul2008.puz", 4},
		{"Jul2410.puz", 4},
		{"Jul3110.puz", 4},
		{"Jun0510.puz", 4},
		{"Jun0610.puz", 6},
		{"Jun1210.puz", 4},
		{"Jun1310.puz", 4},
		{"Jun1910.puz", 4},
		{"Jun2010.puz", 6},
		{"Jun2610.puz", 4},
		{"Jun2710.puz", 6},
		{"Mar0610.puz", 4},
		{"Mar0710.puz", 6},
		{"Mar1008.puz", 4},
		{"Mar1310.puz", 4},
		{"Mar1320.puz", 4},
		{"Mar1410.puz", 6},
		{"Mar1420.puz", 4},
		{"Mar1520.puz", 6},
		{"Mar1620.puz", 4},
		{"Mar1720.puz", 4},
		{"Mar1820.puz", 4},
		{"Mar1920.puz", 4},
		{"Mar2010.puz", 4},
		{"Mar2110.puz", 6},
		{"Mar2710.puz", 4},
		{"Mar2711.puz", 4},
		{"Mar2810.puz", 6},
		{"May0110.puz", 4},
		{"May0210.puz", 6},
		{"May0810.puz", 4},
		{"May0910.puz", 6},
		{"May1510.puz", 4},
		{"May1610.puz", 2},
		{"May2210.puz", 4},
		{"May2310.puz", 6},
		{"May2910.puz", 4},
		{"May3010.puz", 6},
		{"Nov0610.puz", 4},
		{"Nov0710.puz", 6},
		{"Nov1310.puz", 4},
		{"Nov1410.puz", 6},
		{"Nov2010.puz", 4},
		{"Nov2110.puz", 6},
		{"Nov2710.puz", 4},
		{"Nov2810.puz", 6},
		{"Oct0210.puz", 4},
		{"Oct0310.puz", 4},
		{"Oct0910.puz", 4},
		{"Oct1010.puz", 6},
		{"Oct1610.puz", 4},
		{"Oct1710.puz", 4},
		{"Oct1712.puz", 4},
		{"Oct2112.puz", 6},
		{"Oct2310.puz", 4},
		{"Oct2410.puz", 6},
		{"Oct3010.puz", 4},
		{"Oct3110.puz", 6},
		{"Sep0410.puz", 4},
		{"Sep0510.puz", 6},
		{"Sep1108.puz", 4},
		{"Sep1110.puz", 4},
		{"Sep1208.puz", 4},
		{"Sep1210.puz", 6},
		{"Sep1408.puz", 6},
		{"Sep1810.puz", 4},
		{"Sep1908.puz", 6},
		{"Sep1910.puz", 4},
		{"Sep2510.puz", 4},
		{"Sep2610.puz", 6},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			p, err := Read(path.Join(testDataDir, c.file))
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			pdf := gofpdf.New("L", "pt", "Letter", "")
			rc := p.NewRenderContext(pdf)
			pdf.Close()
			n := rc.Layouts[rc.BestLayout].NumColumns
			if n != c.best {
				t.Errorf("got best layout = %d columns, want %d columns", n, c.best)
			}
		})
	}
}
