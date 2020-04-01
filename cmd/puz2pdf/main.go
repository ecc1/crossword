package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/ecc1/crossword"
	"github.com/jung-kurt/gofpdf"
)

var (
	forceFlag  = flag.Bool("f", false, "force overwriting of PDF output file")
	multiFlag  = flag.Bool("m", false, "generate multiple layouts for debugging")
	outputFile = flag.String("o", "", "write PDF output to `file`")
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fail(fmt.Errorf("no input files"))
	}
	pdf, err := setupOutput()
	if err != nil {
		fail(err)
	}
	for _, file := range flag.Args() {
		err := puz2pdf(file, pdf)
		if err != nil {
			fail(err)
		}
	}
	pdf.OutputFileAndClose(*outputFile)
}

func setupOutput() (*gofpdf.Fpdf, error) {
	if *outputFile == "" {
		if flag.NArg() > 1 {
			return nil, fmt.Errorf("multiple input files; output file must be specified with \"-o\"")
		}
		file := flag.Arg(0)
		base := path.Base(file)
		ext := path.Ext(base)
		if ext != ".puz" {
			return nil, fmt.Errorf("%s: file name must end with .puz", file)
		}
		*outputFile = base[:len(base)-len(ext)] + ".pdf"
	}
	if exists(*outputFile) && !*forceFlag {
		return nil, fmt.Errorf("output file %s already exists; use \"-f\" to overwrite", *outputFile)
	}
	return gofpdf.New("L", "pt", "Letter", ""), nil
}

func puz2pdf(file string, pdf *gofpdf.Fpdf) error {
	puz, err := crossword.Read(file)
	if err != nil {
		return err
	}
	rc := puz.NewRenderContext(pdf)
	if *multiFlag {
		rc.RenderAll()
	} else {
		rc.Render()
	}
	return nil
}

func exists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	os.Exit(1)
}
