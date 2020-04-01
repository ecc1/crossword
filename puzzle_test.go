package crossword

import (
	"io/ioutil"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"
)

const testDataDir = "testdata"

func testFiles() []string {
	entries, err := ioutil.ReadDir(testDataDir)
	if err != nil {
		panic(err)
	}
	files := make([]string, len(entries))
	for i, e := range entries {
		files[i] = path.Join(testDataDir, e.Name())
	}
	return files
}

const dateLayout = "Jan0206.puz"

func dateFromFileName(base string) time.Time {
	t, err := time.Parse(dateLayout, base)
	if err != nil {
		return time.Time{}
	}
	return t
}

func TestReadAllPuzzles(t *testing.T) {
	for _, file := range testFiles() {
		base := path.Base(file)
		t.Run(base, func(t *testing.T) {
			p, err := Read(file)
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			date := dateFromFileName(base)
			if date.IsZero() {
				return
			}
			switch date.Weekday() {
			case 0:
				if p.Width != 21 && p.Width != 23 {
					t.Logf("%d×%d Sunday puzzle", p.Width, p.Height)
					return
				}
			default:
				if p.Width != 15 {
					t.Logf("%d×%d weekday puzzle", p.Width, p.Height)
					return
				}
			}
		})
	}
}

func TestReadPuzzle(t *testing.T) {
	cases := []struct {
		file    string
		puzzle  *Puzzle
		circled []Position
	}{
		{
			file:   "Apr2510.puz",
			puzzle: &PuzzleApr2510,
			circled: []Position{
				{8, 0},
				{8, 1},
				{0, 12},
				{20, 13},
				{8, 17},
				{11, 19}, {12, 19},
			},
		},
		{
			file:   "Mar2711.puz",
			puzzle: &PuzzleMar2711,
			circled: []Position{
				{0, 2}, {16, 2},
				{0, 7}, {16, 7},
				{0, 12}, {16, 12},
				{0, 16}, {16, 16},
				{0, 21}, {16, 21},
				{0, 26}, {16, 26},
			},
		},
		{
			file:    "Mar1420.puz",
			puzzle:  &PuzzleMar1420,
			circled: []Position{},
		},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			p, err := Read(path.Join(testDataDir, c.file))
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			if p.Author != c.puzzle.Author {
				t.Errorf("Author == %q, want %q", p.Author, c.puzzle.Author)
			}
			if p.Copyright != c.puzzle.Copyright {
				t.Errorf("Copyright == %q, want %q", p.Copyright, c.puzzle.Copyright)
			}
			if p.Title != c.puzzle.Title {
				t.Errorf("Title == %q, want %q", p.Title, c.puzzle.Title)
			}
			if p.Notepad != c.puzzle.Notepad {
				t.Errorf("Notepad == %q, want %q", p.Notepad, c.puzzle.Notepad)
			}
			if p.Width != c.puzzle.Width {
				t.Errorf("Width == %d, want %d", p.Width, c.puzzle.Width)
			}
			if p.Height != c.puzzle.Height {
				t.Errorf("Height == %d, want %d", p.Height, c.puzzle.Height)
			}
			for i := range p.Dir {
				dir := Direction(i)
				got := &p.Dir[dir]
				want := &c.puzzle.Dir[dir]
				checkNumbers(t, dir, got.Numbers, want.Clues)
				checkMap(t, dir, got.Clues, want.Clues)
				if !p.Scrambled {
					checkMap(t, dir, got.Answers, want.Answers)
				}
				checkIndexes(t, p, dir)
				checkPositions(t, p, dir)
				checkWords(t, p, dir)
			}
			checkCircles(t, p, c.circled)
		})
	}
}

func inPositions(x int, y int, positions []Position) bool {
	for _, pos := range positions {
		if pos.X == x && pos.Y == y {
			return true
		}
	}
	return false
}

func checkCircles(t *testing.T, p *Puzzle, circled []Position) {
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			if inPositions(x, y, circled) {
				if !p.IsCircled(x, y) {
					t.Errorf("square (%d,%d) is not circled but should be", x, y)
				}
			} else {
				if p.IsCircled(x, y) {
					t.Errorf("square (%d,%d) is circled but should not be", x, y)
				}
			}
		}
	}
}

func checkNumbers(t *testing.T, dir Direction, got []int, clues IndexedStrings) {
	want := make([]int, 0, len(got))
	for n := range clues {
		want = append(want, n)
	}
	sort.Ints(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v numbers: got %v, want %v", dir, got, want)
	}
}

func checkMap(t *testing.T, dir Direction, got IndexedStrings, want IndexedStrings) {
	for n, w := range want {
		g := got[n]
		if w != g {
			t.Errorf("%v %d: got %q, want %q", dir, n, g, w)
		}
	}
	for n, g := range got {
		w := want[n]
		if w != "" {
			continue
		}
		t.Errorf("%v %d: got %q (unexpected)", dir, n, g)
	}
}

func checkIndexes(t *testing.T, p *Puzzle, dir Direction) {
	got := &p.Dir[dir]
	for num, i := range got.Indexes {
		n := got.Numbers[i]
		if n != num {
			t.Errorf("Indexes[%d] = %d but Numbers[%d] = %d", num, i, i, n)
		}
	}
}

func checkPositions(t *testing.T, p *Puzzle, dir Direction) {
	got := &p.Dir[dir]
	for num, pos := range got.Positions {
		n := p.PositionNumber(pos)
		if n != num {
			t.Errorf("Position %v for %d %v has number %d", pos, num, dir, n)
		}
		word := got.Words[num]
		if !inPositions(pos.X, pos.Y, word) {
			t.Errorf("Position %v for %d %v is not in word %v", pos, num, dir, word)
		}
	}
}

func checkWords(t *testing.T, p *Puzzle, dir Direction) {
	got := &p.Dir[dir]
	for num, word := range got.Words {
		for _, pos := range word {
			n := int(got.Start[pos.Y][pos.X])
			if n != num {
				t.Errorf("Start%v for %d %v has number %d", pos, num, dir, n)
			}
		}
	}
}

func checkStart(t *testing.T, p *Puzzle, dir Direction) {
	got := &p.Dir[dir]
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			pos := NewPosition(x, y)
			n := int(got.Start[y][x])
			word := got.Words[n]
			if word[0] != pos {
				t.Errorf("Start%v is %d %v but corresponding word starts at %v", pos, n, dir, word[0])
			}
		}
	}
}
