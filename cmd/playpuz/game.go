package main

import (
	"bytes"
	"fmt"

	"github.com/ecc1/crossword"
)

const (
	Across = crossword.Across
	Down   = crossword.Down

	blackSquare = '.'
	emptySquare = ' '
	wrongSquare = '?'
)

var (
	cells        crossword.Grid
	homePos      crossword.Position
	endPos       crossword.Position
	cur          crossword.Position
	curWord      crossword.Word
	curDirection crossword.Direction
)

func initGame() {
	cells = puz.MakeGrid()
	for y := 0; y < puz.Height; y++ {
		cells[y] = make([]byte, puz.Width)
		for x := 0; x < puz.Width; x++ {
			if puz.IsBlack(x, y) {
				cells[y][x] = blackSquare
			} else {
				cells[y][x] = emptySquare
			}
		}
	}
	curDirection = Across
	d := &puz.Dir[curDirection]
	homePos = d.Positions[1]
	cur = homePos
	lastNum := d.Numbers[len(d.Numbers)-1]
	endPos = d.Positions[lastNum]
}

func getContents() []byte {
	return cells.Contents()
}

func setContents(contents []byte) error {
	contents = bytes.ReplaceAll(contents, []byte{'\n'}, nil)
	if len(contents) != puz.Width*puz.Height {
		return fmt.Errorf("contents do not match this puzzle")
	}
	for y := 0; y < puz.Height; y++ {
		for x := 0; x < puz.Width; x++ {
			if !puz.IsBlack(x, y) {
				cells[y][x] = contents[0]
				redrawCell(x, y)
			}
			contents = contents[1:]
		}
	}
	return nil
}

func puzzleIsSolved() bool {
	for y := 0; y < puz.Height; y++ {
		for x := 0; x < puz.Width; x++ {
			if puz.IsBlack(x, y) {
				continue
			}
			if cells[y][x] != puz.Answer(x, y) {
				return false
			}
		}
	}
	return true
}

func moveHome() {
	setActivePos(homePos)
}

func moveEnd() {
	setActivePos(endPos)
}

func moveLeft() {
	if curDirection == Down {
		changeDirection()
		return
	}
	moveBackward(false)
}

func moveRight() {
	if curDirection == Down {
		changeDirection()
		return
	}
	moveForward(false)
}

func moveUp() {
	if curDirection == Across {
		changeDirection()
		return
	}
	moveBackward(false)
}

func moveDown() {
	if curDirection == Across {
		changeDirection()
		return
	}
	moveForward(false)
}

func updateCell(c uint) {
	cells[cur.Y][cur.X] = byte(c)
	redrawCell(cur.X, cur.Y)
	if puzzleIsSolved() {
		winnerWinner()
	}
}

func changeDirection() {
	curDirection = 1 - curDirection
	setActivePos(cur)
}

func setActivePos(pos crossword.Position) {
	cur = pos
	d := &puz.Dir[curDirection]
	oldWord := curWord
	num := int(d.Start[cur.Y][cur.X])
	if num == 0 {
		// No word in this direction.
		return
	}
	curWord = d.Words[num]
	redrawWord(oldWord)
	redrawWord(curWord)
	highlightClues()
}

func setActive(x, y int) {
	setActivePos(crossword.NewPosition(x, y))
}

func redrawWord(word crossword.Word) {
	for _, pos := range word {
		redrawCell(pos.X, pos.Y)
	}
}

func isActive(x, y int) bool {
	return x == cur.X && y == cur.Y
}

func inActiveWord(x, y int) bool {
	return positionInWord(crossword.NewPosition(x, y), curWord) != -1
}

func positionInWord(pos crossword.Position, word crossword.Word) int {
	for i, p := range word {
		if p == pos {
			return i
		}
	}
	return -1
}

func moveForward(skip bool) {
	d := &puz.Dir[curDirection]
	num := int(d.Start[cur.Y][cur.X])
	if num == 0 {
		// No word in this direction.
		return
	}
	word := d.Words[num]
	i := positionInWord(cur, word)
	if i == -1 {
		panic("moveForward")
	}
	for i < len(word)-1 {
		i++
		pos := word[i]
		if !skip || cells[pos.Y][pos.X] == emptySquare {
			setActivePos(pos)
			return
		}
	}
	// Find the next empty square, if any.
	for k := d.Indexes[num] + 1; k < len(d.Numbers); k++ {
		num := d.Numbers[k]
		word := d.Words[num]
		for _, pos := range word {
			if cells[pos.Y][pos.X] == emptySquare {
				setActivePos(pos)
				return
			}
		}
		// This word is all filled in, try the next.
	}
}

func moveBackward(skip bool) {
	d := &puz.Dir[curDirection]
	num := int(d.Start[cur.Y][cur.X])
	if num == 0 {
		// No word in this direction.
		return
	}
	word := d.Words[num]
	i := positionInWord(cur, word)
	if i == -1 {
		panic("moveBackward")
	}
	for i > 0 {
		i--
		pos := word[i]
		if !skip || cells[pos.Y][pos.X] == emptySquare {
			setActivePos(pos)
			return
		}
	}
	k := d.Indexes[num]
	// If not on the first word, move to the end of the previous one.
	if k > 0 {
		prevNum := d.Numbers[k-1]
		prevWord := d.Words[prevNum]
		setActivePos(prevWord[len(prevWord)-1])
	}
}

func activateFromClue(n int) {
	pos := puz.Dir[curDirection].Positions[n]
	setActive(pos.X, pos.Y)
}

func highlightClues() {
	for dir, d := range puz.Dir {
		num := int(d.Start[cur.Y][cur.X])
		if num == 0 {
			// No word in this direction.
			continue
		}
		i := d.Indexes[num]
		selectClue(crossword.Direction(dir), i)
	}
}

func checkSquare(x, y int) {
	c := cells[y][x]
	if c == emptySquare || c == puz.Answer(x, y) {
		return
	}
	cells[y][x] = wrongSquare
	redrawCell(x, y)
}

func checkWord() {
	for _, pos := range curWord {
		checkSquare(pos.X, pos.Y)
	}
}

func checkPuzzle() {
	for y := 0; y < puz.Height; y++ {
		for x := 0; x < puz.Width; x++ {
			if puz.IsBlack(x, y) {
				continue
			}
			checkSquare(x, y)
		}
	}
}

func solveWord() {
	for _, pos := range curWord {
		x, y := pos.X, pos.Y
		cells[y][x] = puz.Answer(x, y)
		redrawCell(x, y)
	}
}

func solvePuzzle() {
	setContents(puz.SolutionBytes())
}
