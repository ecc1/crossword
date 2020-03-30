package acrosslite

import (
	"bytes"
	"fmt"
)

func (p *Puzzle) Unlock() (int, error) {
	if !p.Scrambled {
		return 0, nil
	}
	src := p.compressBuffer(p.solution)
	dst := make([]byte, len(src))
	tmp := make([]byte, len(src))
	key := make([]uint8, 4)
	for nextKey(key) {
		unscramble(src, key, dst, tmp)
		if p.correctAnswers(dst) {
			p.solution = p.expandBuffer(dst)
			p.Scrambled = false
			return keyNumber(key), nil
		}
	}
	return 0, fmt.Errorf("brute-force unlocking failed")
}

func nextKey(key []uint8) bool {
	for i := range key {
		key[i]++
		if key[i] != 10 {
			return true
		}
		key[i] = 0
	}
	return false
}

func keyNumber(key []uint8) int {
	n := 0
	for i := range key {
		n = 10*n + int(key[i])
	}
	return n
}

func (p *Puzzle) UnlockWithKey(key int) error {
	if !p.Scrambled {
		if key == 0 {
			return nil
		}
		return fmt.Errorf("puzzle is already unlocked")
	}
	if key < 0000 || 9999 < key {
		return fmt.Errorf("key (%d) must be in the range 0000 .. 9999", key)
	}
	// Convert digits to corresponding ints.
	k := []uint8(fmt.Sprintf("%04d", key))
	for i, b := range k {
		k[i] = b - '0'
	}
	src := p.compressBuffer(p.solution)
	dst := make([]byte, len(src))
	tmp := make([]byte, len(src))
	unscramble(src, k, dst, tmp)
	if !p.correctAnswers(dst) {
		return fmt.Errorf("key %04d does not unlock this puzzle", key)
	}
	p.solution = p.expandBuffer(dst)
	p.Scrambled = false
	return nil
}

func unscramble(src []byte, key []uint8, dst []byte, tmp []byte) {
	n := len(src)
	copy(dst, src)
	for i := range key {
		unshuffle(dst, tmp)
		k := int(key[3-i])
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

func unshift(buf []byte, key []uint8) {
	for i, c := range buf {
		buf[i] = 'A' + ((c-'A')-key[i%4]+26)%26
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
