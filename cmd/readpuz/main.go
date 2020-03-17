package main

import (
	"fmt"
	"os"

	"github.com/ecc1/acrosslite"
)

func main() {
	p, err := acrosslite.Read(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", *p)
}
