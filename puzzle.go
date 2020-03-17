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
		Scrambled bool
		Clues     []string

		AcrossClues   IndexedStrings
		AcrossAnswers IndexedStrings
		DownClues     IndexedStrings
		DownAnswers   IndexedStrings

		// Height * Width grids.
		Numbers  [][]int
		Circles  Grid
		Solution Grid
	}

	Checksums struct {
		Global    uint16
		Scrambled uint16
		Magic     uint64
	}

	IndexedStrings map[int]string

	Grid [][]byte
)

func (p *Puzzle) IsBlack(x int, y int) bool {
	if x < 0 || x >= p.Width || y < 0 || y >= p.Height {
		return true
	}
	return p.Solution[y][x] == '.'
}

func (p *Puzzle) IsCircled(x int, y int) bool {
	if len(p.Circles) == 0 {
		return false
	}
	return p.Circles[y][x] == 0x80
}

func (p *Puzzle) CellNumber(x int, y int) int {
	return int(p.Numbers[y][x])
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
	p.Solution, puz, err = p.readGrid(puz)
	if err != nil {
		return nil, fmt.Errorf("malformed solution section in %d×%d puzzle: %s", w, h, err)
	}
	if !p.Scrambled {
		n := p.indexAnswers()
		if n != p.NumClues {
			return nil, fmt.Errorf("%d answers were indexed instead of %d", n, p.NumClues)
		}
	}
	_, puz, err = p.readGrid(puz)
	if err != nil {
		return nil, fmt.Errorf("malformed fill section in %d×%d puzzle: %s", w, h, err)
	}

	p.Title, puz = readString(puz)
	p.Author, puz = readString(puz)
	p.Copyright, puz = readString(puz)

	p.Numbers = p.gridNumbers()
	p.Clues = make([]string, p.NumClues)
	for i := range p.Clues {
		p.Clues[i], puz = readString(puz)
	}
	n := p.indexClues()
	if n != p.NumClues {
		return nil, fmt.Errorf("%d clues were indexed instead of %d", n, p.NumClues)
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

func origBytes(s string) []byte {
	v, _ := charmap.Windows1252.NewEncoder().Bytes([]byte(s))
	return v
}

func (p *Puzzle) gridNumbers() [][]int {
	g := make([][]int, p.Height)
	for i := range g {
		g[i] = make([]int, p.Width)
	}
	return g
}

func (p *Puzzle) readGrid(v []byte) (Grid, []byte, error) {
	if len(v) < p.Height*p.Width {
		return nil, nil, fmt.Errorf("only %d bytes of grid data instead of %d", len(v), p.Height*p.Width)
	}
	g := make(Grid, p.Height)
	for i := range g {
		g[i] = v[:p.Width]
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
		p.Circles, _, err = p.readGrid(data)
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

// indexAnswers determines answer numbers and adds the corresponding answers to the AcrossAnswers and DownAnswers maps.
// It returns the total number of answers indexed.
func (p *Puzzle) indexAnswers() int {
	p.AcrossAnswers = make(IndexedStrings)
	p.DownAnswers = make(IndexedStrings)
	c := 0 // answer count
	n := 1 // square number
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			if p.IsBlack(x, y) {
				continue
			}
			numbered := false
			if p.IsBlack(x-1, y) && !p.IsBlack(x+1, y) {
				p.AcrossAnswers[n] = p.readAcrossAnswer(x, y)
				numbered = true
				c++
			}
			if p.IsBlack(x, y-1) && !p.IsBlack(x, y+1) {
				p.DownAnswers[n] = p.readDownAnswer(x, y)
				numbered = true
				c++
			}
			if numbered {
				n++
			}
		}
	}
	return c
}

func (p *Puzzle) readAcrossAnswer(x int, y int) string {
	var sb strings.Builder
	for i := x; i < p.Width; i++ {
		if p.IsBlack(i, y) {
			return sb.String()
		}
		sb.WriteByte(p.Solution[y][i])
	}
	return sb.String()
}

func (p *Puzzle) readDownAnswer(x int, y int) string {
	var sb strings.Builder
	for j := y; j < p.Height; j++ {
		if p.IsBlack(x, j) {
			return sb.String()
		}
		sb.WriteByte(p.Solution[j][x])
	}
	return sb.String()
}

// indexClues determines clue numbers, adds the corresponding clues to the AcrossClues and DownClues maps,
// and enters the number in the Numbers grid. It returns the total number of clues indexed.
func (p *Puzzle) indexClues() int {
	p.AcrossClues = make(IndexedStrings)
	p.DownClues = make(IndexedStrings)
	c := 0 // clue index
	n := 1 // square number
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			if p.IsBlack(x, y) {
				continue
			}
			numbered := false
			if p.IsBlack(x-1, y) && !p.IsBlack(x+1, y) {
				p.AcrossClues[n] = p.Clues[c]
				numbered = true
				c++
			}
			if p.IsBlack(x, y-1) && !p.IsBlack(x, y+1) {
				p.DownClues[n] = p.Clues[c]
				numbered = true
				c++
			}
			if numbered {
				p.Numbers[y][x] = n
				n++
			}
		}
	}
	return c
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
	for _, clue := range p.Clues {
		c = stringChecksum(clue, c)
	}
	if p.Version >= "1.3" {
		c = zStringChecksum(p.Notepad, c)
	}
	return c
}

func stringChecksum(s string, c uint16) uint16 {
	return checksum(origBytes(s), c)
}

func zStringChecksum(s string, c uint16) uint16 {
	if s == "" {
		return c
	}
	return stringChecksum(s+"\x00", c)
}
