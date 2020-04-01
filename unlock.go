package crossword

import (
	"bytes"
	"fmt"
)

type (
	// Key stores a 4-digit decimal key in big-endian order, one digit per byte.
	Key []uint8
)

func NewKey() Key {
	return make(Key, 4)
}

func NewKeyFromInt(k int) (Key, error) {
	key := NewKey()
	m := k
	for i := 3; i >= 0; i-- {
		if m <= 0 {
			break
		}
		key[i] = uint8(m % 10)
		m /= 10
	}
	if m != 0 {
		return nil, fmt.Errorf("key (%d) must be in the range 0000 .. 9999", k)
	}
	return key, nil
}

// Next increments key in place, and returns true if it did not overflow.
func (key Key) Next() bool {
	for i := 3; i >= 0; i-- {
		key[i]++
		if key[i] != 10 {
			return true
		}
		// Carry into next digit.
		key[i] = 0
	}
	return false
}

func (key Key) Int() int {
	n := 0
	for _, v := range key {
		n = 10*n + int(v)
	}
	return n
}

func (p *Puzzle) Unlock() (int, error) {
	if !p.Scrambled {
		return 0, nil
	}
	src := p.compressBuffer(p.solution)
	dst := make([]byte, len(src))
	tmp := make([]byte, len(src))
	key := NewKey()
	for {
		unscramble(src, key, dst, tmp)
		if p.correctAnswers(dst) {
			p.solution = p.expandBuffer(dst)
			p.Scrambled = false
			return key.Int(), nil
		}
		if !key.Next() {
			break
		}
	}
	return 0, fmt.Errorf("brute-force unlocking failed")
}

func (p *Puzzle) UnlockWithKey(k int) error {
	if !p.Scrambled {
		if k == 0 {
			return nil
		}
		return fmt.Errorf("puzzle is already unlocked")
	}
	key, err := NewKeyFromInt(k)
	if err != nil {
		return err
	}
	src := p.compressBuffer(p.solution)
	dst := make([]byte, len(src))
	tmp := make([]byte, len(src))
	unscramble(src, key, dst, tmp)
	if !p.correctAnswers(dst) {
		return fmt.Errorf("key %04d does not unlock this puzzle", k)
	}
	p.solution = p.expandBuffer(dst)
	p.Scrambled = false
	return nil
}

func unscramble(src []byte, key Key, dst []byte, tmp []byte) {
	n := len(src)
	copy(dst, src)
	for i := 3; i >= 0; i-- {
		unshuffle(dst, tmp)
		k := int(key[i])
		copy(dst, tmp[n-k:])
		copy(dst[k:], tmp[:n-k])
		unshift(dst, key)
	}
}

func unshuffle(src []byte, dst []byte) {
	n := len(src)
	j := 0
	for i := 1; i < n; i += 2 {
		dst[j] = src[i]
		j++
	}
	for i := 0; i < n; i += 2 {
		dst[j] = src[i]
		j++
	}
}

func unshift(buf []byte, key Key) {
	for i, c := range buf {
		buf[i] = 'A' + (26+(c-'A')-key[i%4])%26
	}
}

func (p *Puzzle) compressBuffer(g Grid) []byte {
	var buf bytes.Buffer
	for x := 0; x < p.Width; x++ {
		for y := 0; y < p.Height; y++ {
			if p.IsBlack(x, y) {
				continue
			}
			buf.WriteByte(g[y][x])
		}
	}
	return buf.Bytes()
}

func (p *Puzzle) expandBuffer(buf []byte) Grid {
	g := p.MakeGrid()
	for x := 0; x < p.Width; x++ {
		for y := 0; y < p.Height; y++ {
			if p.IsBlack(x, y) {
				g[y][x] = blackSquare
			} else {
				g[y][x] = buf[0]
				buf = buf[1:]
			}
		}
	}
	return g
}

func (p *Puzzle) correctAnswers(buf []byte) bool {
	return checksum(buf, 0) == p.Checksum.Scrambled
}
