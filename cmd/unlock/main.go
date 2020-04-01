package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ecc1/crossword"
)

func main() {
	var file string
	var key int
	var err error
	switch len(os.Args) {
	case 2:
		file = os.Args[1]
	case 3:
		file = os.Args[1]
		key, err = strconv.Atoi(os.Args[2])
		if err != nil {
			fail(err)
		}
		if key == 0 {
			fail(fmt.Errorf("0000 is not a valid key"))
		}
	default:
		fail(fmt.Errorf("Usage: %s file.puz [key]", os.Args[0]))
	}
	puz, err := crossword.Read(file)
	if err != nil {
		fail(err)
	}
	if key != 0 {
		err = puz.UnlockWithKey(key)
	} else {
		key, err = puz.Unlock()
		if err == nil {
			fmt.Printf("[key = %04d]\n", key)
		}
	}
	if err != nil {
		fail(err)
	}
	fmt.Print(puz.Solution())
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	os.Exit(1)
}
