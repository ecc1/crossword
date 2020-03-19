package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ecc1/acrosslite"
)

var (
	goFormat = flag.Bool("g", false, "print puzzle in Go struct format")
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fail(fmt.Errorf("single PUZ file required"))
	}
	p, err := acrosslite.Read(flag.Arg(0))
	if err != nil {
		fail(err)
	}
	if *goFormat {
		fmt.Printf("%#v\n", *p)
	} else {
		fmt.Printf("%+v\n", *p)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	os.Exit(1)
}
