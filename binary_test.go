package crossword

import (
	"fmt"
	"io"
	"math"
	"strings"
	"testing"
)

func TestRead16(t *testing.T) {
	cases := []struct {
		val uint16
		rep []byte
	}{
		{0xCDAB, parseBytes("AB CD")},
		{math.MaxUint16, parseBytes("FF FF")},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c.val), func(t *testing.T) {
			val := read16(c.rep)
			if val != c.val {
				t.Errorf("read16(% X) == %04X, want %04X", c.rep, val, c.val)
			}
		})
	}
}

func TestRead32(t *testing.T) {
	cases := []struct {
		val uint32
		rep []byte
	}{
		{0x67452301, parseBytes("01 23 45 67")},
		{math.MaxUint32, parseBytes("FF FF FF FF")},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c.val), func(t *testing.T) {
			val := read32(c.rep)
			if val != c.val {
				t.Errorf("read32(% X) == %08X, want %08X", c.rep, val, c.val)
			}
		})
	}
}

func TestRead64(t *testing.T) {
	cases := []struct {
		val uint64
		rep []byte
	}{
		{0xEFCDAB8967452301, parseBytes("01 23 45 67 89 AB CD EF")},
		{math.MaxUint64, parseBytes("FF FF FF FF FF FF FF FF")},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c.val), func(t *testing.T) {
			val := read64(c.rep)
			if val != c.val {
				t.Errorf("read64(% X) == %08X, want %08X", c.rep, val, c.val)
			}
		})
	}
}

func readBytes(r io.Reader) ([]byte, error) {
	var data []byte
	for {
		var b byte
		n, err := fmt.Fscanf(r, "%02x", &b)
		if n == 0 {
			break
		}
		if err != nil {
			return data, err
		}
		data = append(data, b)
	}
	return data, nil
}

func parseBytes(hex string) []byte {
	var data []byte
	r := strings.NewReader(hex)
	data, err := readBytes(r)
	if err != nil {
		panic(err)
	}
	return data
}
