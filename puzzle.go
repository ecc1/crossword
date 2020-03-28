package acrosslite

// AcrossLite crossword puzzle reader,
// based on https://github.com/alexdej/puzpy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/bits"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

type (
	Puzzle struct {
		Header   []byte
		Checksum Checksums

		Version string

		Author    string
		Copyright string
		Title     string
		Notepad   string

		Width  int
		Height int

		NumClues  int
		AllClues  []string
		Scrambled bool

		// Clue information indexed by direction.
		Dir []Clue

		// Height * Width grids.
		numbers  Grid
		solution Grid
		circles  Grid
	}

	Direction int

	// Per-direction clue information.
	Clue struct {
		// Clue numbers in increasing order.
		Numbers []int
		// Reverse map from clue number to index in Numbers.
		Indexes map[int]int
		// Clues indexed by their number.
		Clues IndexedStrings
		// Answers indexed by their number.
		Answers IndexedStrings
		// Positions[n] is the position of the square with number n.
		Positions IndexedPositions
		// Words[n] is the Word for answer number n, in left-right or top-down order.
		Words IndexedWords
		// Start[y][x] is clue number for the word that passes through square (x, y).
		// May be zero for uncrossed words in unusual puzzles.
		Start Grid
	}

	Grid [][]uint8

	Position struct {
		X int
		Y int
	}

	Word []Position

	IndexedStrings   map[int]string
	IndexedPositions map[int]Position
	IndexedWords     map[int]Word

	Checksums struct {
		Global    uint16
		Scrambled uint16
		Magic     uint64
	}
)

const (
	Across Direction = 0
	Down   Direction = 1
)

func (dir Direction) String() string {
	switch dir {
	case Across:
		return "ACROSS"
	case Down:
		return "DOWN"
	}
	panic(fmt.Sprintf("Direction %d", dir))
}

func NewPosition(x, y int) Position {
	return Position{X: x, Y: y}
}

func (pos Position) String() string {
	return fmt.Sprintf("(%d, %d)", pos.X, pos.Y)
}

func (p *Puzzle) IsBlack(x, y int) bool {
	if x < 0 || x >= p.Width || y < 0 || y >= p.Height {
		return true
	}
	return p.solution[y][x] == '.'
}

func (p *Puzzle) IsCircled(x, y int) bool {
	if len(p.circles) == 0 {
		return false
	}
	return p.circles[y][x] == 0x80
}

// PositionNumber(pos) is the number for square at position pos, or 0.
func (p *Puzzle) PositionNumber(pos Position) int {
	return p.SquareNumber(pos.X, pos.Y)
}

// SquareNumber(x, y) is the number for square (x, y), or 0.
func (p *Puzzle) SquareNumber(x, y int) int {
	return int(p.numbers[y][x])
}

func Read(file string) (*Puzzle, error) {
	puz, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	p, err := Decode(puz)
	if err != nil {
		err = fmt.Errorf("%s: %s", file, err)
	}
	return p, err
}

func Decode(puz []byte) (*Puzzle, error) {
	var p Puzzle
	var err error
	puz, err = p.readHeader(puz)
	if err != nil {
		return nil, err
	}

	w, h := p.Width, p.Height
	grids := puz // remember start of solution and fill grids for global checksum
	p.solution, puz, err = p.readGrid(puz)
	if err != nil {
		return nil, fmt.Errorf("malformed solution section in %d×%d puzzle: %s", w, h, err)
	}
	_, puz, err = p.readGrid(puz)
	if err != nil {
		return nil, fmt.Errorf("malformed fill section in %d×%d puzzle: %s", w, h, err)
	}

	p.Title, puz = readString(puz)
	p.Author, puz = readString(puz)
	p.Copyright, puz = readString(puz)

	p.numbers = p.makeGrid()
	p.AllClues = make([]string, p.NumClues)
	for i := range p.AllClues {
		p.AllClues[i], puz = readString(puz)
	}
	p.indexClues()
	numAcross := len(p.Dir[Across].Numbers)
	numDown := len(p.Dir[Down].Numbers)
	n := numAcross + numDown
	if n != p.NumClues {
		return nil, fmt.Errorf("%d %v + %d %v clues were indexed instead of %d", numAcross, Across, numDown, Down, p.NumClues)
	}

	p.Notepad, puz = readString(puz)

	err = p.validateChecksums(grids)
	if err != nil {
		return nil, err
	}

	for {
		puz, err = p.readExtension(puz)
		if err != nil {
			return nil, fmt.Errorf("malformed extension section in %d×%d puzzle: %s", w, h, err)
		}
		if puz == nil {
			break
		}
	}

	return &p, nil
}

const headerLength = 52

var magic = []byte("ACROSS&DOWN\x00")

func (p *Puzzle) readHeader(v []byte) ([]byte, error) {
	if len(v) < headerLength {
		return nil, fmt.Errorf("puzzle is only %d bytes long", len(v))
	}
	// Header starts 2 bytes before the "ACROSS&DOWN" string.
	i := bytes.Index(v, magic) - 2
	if i < 0 {
		return nil, fmt.Errorf("puzzle does not contain expected header %q", magic)
	}
	headerEnd := i + headerLength
	p.Header = v[i:headerEnd]
	h := p.Header
	check := read16(h[14:16])
	calc := p.headerChecksum()
	if calc != check {
		return nil, fmt.Errorf("header checksum = %04X, expected %04X", calc, check)
	}
	p.Checksum = Checksums{
		Global:    read16(h[0:2]),
		Magic:     read64(h[16:24]),
		Scrambled: read16(h[30:32]),
	}
	p.Version = string(h[24:27])
	p.Width = int(h[44])
	p.Height = int(h[45])
	p.NumClues = int(read16(h[46:48]))
	p.Scrambled = read16(h[49:51]) != 0
	return v[headerEnd:], nil
}

func read16(data []byte) uint16 {
	return uint16(data[1])<<8 | uint16(data[0])
}

func read32(data []byte) uint32 {
	return uint32(read16(data[2:4]))<<16 | uint32(read16(data[0:2]))
}

func read64(data []byte) uint64 {
	return uint64(read32(data[4:8]))<<32 | uint64(read32(data[0:4]))
}

func readString(v []byte) (string, []byte) {
	var buf strings.Builder
	for i, b := range v {
		if b == 0 {
			return makeString(buf.String()), v[i+1:]
		}
		buf.WriteByte(b)
	}
	return makeString(buf.String()), nil
}

func makeString(orig string) string {
	s, _ := charmap.Windows1252.NewDecoder().String(orig)
	return s
}

func PuzzleBytes(s string) []byte {
	v, _ := charmap.Windows1252.NewEncoder().Bytes([]byte(s))
	return v
}

func (p *Puzzle) makeGrid() Grid {
	g := make(Grid, p.Height)
	for i := range g {
		g[i] = make([]uint8, p.Width)
	}
	return g
}

func (p *Puzzle) readGrid(v []byte) (Grid, []byte, error) {
	if len(v) < p.Height*p.Width {
		return nil, nil, fmt.Errorf("only %d bytes of grid data instead of %d", len(v), p.Height*p.Width)
	}
	g := p.makeGrid()
	for i := range g {
		copy(g[i], v[:p.Width])
		v = v[p.Width:]
	}
	return g, v, nil
}

func (p *Puzzle) readExtension(v []byte) ([]byte, error) {
	if len(v) < 8 {
		return nil, nil
	}
	code := string(v[0:4])
	count := int(read16(v[4:6]))
	check := read16(v[6:8])
	v = v[8:]
	if len(v) < count+1 || count < 0 {
		return nil, fmt.Errorf("only %d bytes of %s extension data instead of %d", len(v), code, count+1)
	}
	data := v[:count]
	calc := checksum(data, 0)
	if calc != check {
		return nil, fmt.Errorf("%s extension checksum = %04X, expected %04X", code, calc, check)
	}
	switch code {
	case "GEXT":
		if len(data) != p.Height*p.Width {
			return nil, fmt.Errorf("%s extension contains %d bytes of data instead of %d", code, len(data), p.Height*p.Width)
		}
		var err error
		p.circles, _, err = p.readGrid(data)
		if err != nil {
			return nil, fmt.Errorf("%s extension: %s", code, err)
		}
	case "GRBS":
	case "LTIM":
	case "RTBL":
	case "RUSR":
	default:
		return nil, fmt.Errorf("unsupported %s extension", code)
	}
	return v[count+1:], nil
}

// indexClues determines clue numbers and indexes their positions, numbers, clues, and answers.
func (p *Puzzle) indexClues() {
	p.Dir = make([]Clue, 2)
	for i := range p.Dir {
		d := &p.Dir[i]
		d.Indexes = make(map[int]int)
		d.Clues = make(IndexedStrings)
		d.Answers = make(IndexedStrings)
		d.Positions = make(IndexedPositions)
		d.Words = make(IndexedWords)
		d.Start = p.makeGrid()
	}
	c := 0 // clue index
	n := 1 // square number
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			if p.IsBlack(x, y) {
				continue
			}
			numbered := false
			if p.IsBlack(x-1, y) && !p.IsBlack(x+1, y) {
				p.readAnswer(n, Across, x, y, p.AllClues[c])
				numbered = true
				c++
			}
			if p.IsBlack(x, y-1) && !p.IsBlack(x, y+1) {
				p.readAnswer(n, Down, x, y, p.AllClues[c])
				numbered = true
				c++
			}
			if numbered {
				p.numbers[y][x] = uint8(n)
				n++
			}
		}
	}
}

func (p *Puzzle) readAnswer(n int, dir Direction, x, y int, clue string) {
	d := &p.Dir[dir]
	var sb strings.Builder
	var word Word
	switch dir {
	case Across:
		for i := x; i < p.Width; i++ {
			if p.IsBlack(i, y) {
				break
			}
			word = append(word, NewPosition(i, y))
			sb.WriteByte(p.solution[y][i])
			d.Start[y][i] = uint8(n)
		}
	case Down:
		for j := y; j < p.Height; j++ {
			if p.IsBlack(x, j) {
				break
			}
			word = append(word, NewPosition(x, j))
			sb.WriteByte(p.solution[j][x])
			d.Start[j][x] = uint8(n)
		}
	}
	answer := sb.String()
	d.Positions[n] = NewPosition(x, y)
	d.Numbers = append(d.Numbers, n)
	d.Indexes[n] = len(d.Numbers) - 1
	d.Clues[n] = clue
	d.Answers[n] = answer
	d.Words[n] = word
}

func checksum(data []byte, c uint16) uint16 {
	for _, b := range data {
		c = bits.RotateLeft16(c, -1) + uint16(b)
	}
	return c
}

var mask = []byte("ICHEATED")

func (p *Puzzle) validateChecksums(grids []byte) error {
	n := p.Height * p.Width

	c := p.headerChecksum()
	c = checksum(grids[:2*n], c)
	c = p.textChecksum(c)
	if c != p.Checksum.Global {
		return fmt.Errorf("global checksum = %04X, expected %04X", c, p.Checksum.Global)
	}

	sums := []uint16{
		p.textChecksum(0),
		checksum(grids[n:2*n], 0),
		checksum(grids[:n], 0),
		p.headerChecksum(),
	}
	m := uint64(0)
	for i, c := range sums {
		m <<= 8
		m |= uint64(mask[7-i]^uint8(c>>8)) << 32
		m |= uint64(mask[3-i]^uint8(c)) << 0
	}
	if m != p.Checksum.Magic {
		return fmt.Errorf("magic checksum = %016X, expected %016X", c, p.Checksum.Magic)
	}

	return nil
}

func (p *Puzzle) headerChecksum() uint16 {
	return checksum(p.Header[44:52], 0)
}

func (p *Puzzle) textChecksum(c uint16) uint16 {
	c = zStringChecksum(p.Title, c)
	c = zStringChecksum(p.Author, c)
	c = zStringChecksum(p.Copyright, c)
	for _, clue := range p.AllClues {
		c = stringChecksum(clue, c)
	}
	if p.Version >= "1.3" {
		c = zStringChecksum(p.Notepad, c)
	}
	return c
}

func stringChecksum(s string, c uint16) uint16 {
	return checksum(PuzzleBytes(s), c)
}

func zStringChecksum(s string, c uint16) uint16 {
	if s == "" {
		return c
	}
	return stringChecksum(s+"\x00", c)
}
