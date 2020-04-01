package main

import (
	"fmt"
	"os"

	"github.com/ecc1/crossword"
)

var (
	puz *crossword.Puzzle
)

func main() {
	if len(os.Args) != 2 {
		fail(fmt.Errorf("single PUZ file required"))
	}
	var err error
	puz, err = crossword.Read(os.Args[1])
	if err != nil {
		fail(err)
	}
	if puz.Scrambled {
		key, err := puz.Unlock()
		if err != nil {
			fail(err)
		}
		fmt.Printf("unlocked puzzle with key %04d\n", key)
	}
	initGame()
	initUI()
	runUI()
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	os.Exit(1)
}
