package crossword

import (
	"bytes"
	"fmt"
	"path"
	"testing"
)

var unlockCases = []struct {
	file     string
	key      int
	locked   string
	unlocked string
}{
	{
		file: "Apr2510.puz",
		key:  4462,
		unlocked: `FATSO.KAPUT.SCAD.EDAM
ASHER.OMANI.ALEE.AMTO
CHEERLEADINGFORMATION
TOSS.ANDS..RENOIR.TNT
STEFANI..SKICAP.AURAE
...IMAGEONADOLLARBILL
OFFTO..DOIT...ALAS...
ROI..TIGHT.ASKNOT.CTS
YOGAPOSE..SLICES.DAWS
.DTRAIN.HOTTEA.SLIPON
APRES.THELOUVRE.OMITS
CLEATS.UNEASE.STRATI.
MAES.AMNIOS..CARDGAME
ENS.FLAKES.AFOUL..LEY
...LOAD...AMES..BLARE
GLASSDESIGNBYIMPEI...
LULUS.NUCLEI..AUSTRIA
USA.ANOMIE..CMDR.CONK
THREEDIMENSIONALSHAPE
ELMS.ASES.CDROM.RIDER
SYST.KERT.HEAPS.ASSNS
`,
	},
	{
		file: "Aug0810.puz",
		key:  5622,
		unlocked: `FABIAN..RETRO.MADEPAR
APOLLO.TEGRIN.OBVERSE
TELEPATHSGIFT.BARRONS
AXE.HIREON.FOLIC.SMEE
HERR.RAMROD..ILIE.IRT
.SOAS.WEBGIGGLE.MESS.
...SPILT..LETT.DBLS..
MEETERS.POLES.REASON.
ALSACE.CAPOS.SOAR.REA
RAPS..LOSTNETWORK.YSL
CSI.PIOUS...AAMES.NTH
ITO.RNSPECIALTY..SOLI
AIN.ONES.ASICS.PLATER
.CASTER.STATS.SOONEST
..GEER.BOCA..CHILD...
.MEWS.CASHCACHE.LANS.
FOG.TEAR..STAIRS.LACK
INRE.STEAD.OPERAS.MAY
DIORITE.CRIMEFIGHTERS
ECUADOR.TAPIRS.EUNICE
LAPTOPS.IMACS..SITTER
`,
	},
	{
		file: "May1510.puz",
		key:  9798,
		unlocked: `DUPLICATORS.AWE
ONLINEMEDIA.PEA
STAGFLATION.PET
TENNESSEETITANS
.STIR.....TALIA
STETSONS.VARLET
PERE.VIEWERS...
ADS.PEPPERY.BET
...TONSILS.BASE
BLOATS.ALOUETTE
RANCH.....NAPE.
AUTHORPUBLISHER
IRA.LEAVEITTOME
DIP.ENTERTAINED
SEE.SESAMESEEDS
`,
	},
}

func TestUnlockWithKey(t *testing.T) {
	for _, c := range unlockCases {
		t.Run(c.file, func(t *testing.T) {
			p, err := Read(path.Join(testDataDir, c.file))
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			err = p.UnlockWithKey(c.key)
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			if p.Scrambled {
				t.Errorf("Scrambled flag is still true")
			}
			if p.Solution() != c.unlocked {
				t.Errorf("unlocked solution is not correct")
			}
		})
	}
}

func TestUnlock(t *testing.T) {
	for _, c := range unlockCases {
		t.Run(c.file, func(t *testing.T) {
			p, err := Read(path.Join(testDataDir, c.file))
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			key, err := p.Unlock()
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			if p.Scrambled {
				t.Errorf("Scrambled flag is still true")
			}
			if p.Solution() != c.unlocked {
				t.Errorf("unlocked solution is not correct")
			}
			if key != c.key {
				t.Errorf("brute-force unlock found key %04d, want %04d", key, c.key)
			}
		})
	}
}

func TestUnlockAllPuzzles(t *testing.T) {
	for _, file := range testFiles() {
		base := path.Base(file)
		t.Run(base, func(t *testing.T) {
			p, err := Read(file)
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			if !p.Scrambled {
				return
			}
			key, err := p.Unlock()
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			if p.Scrambled {
				t.Errorf("Scrambled flag is still true")
			}
			t.Logf("key = %04d", key)
		})
	}
}

const (
	MaxInt = int(^uint(0) >> 1)
	MinInt = int(-MaxInt - 1)
)

func TestKeys(t *testing.T) {
	errorCase := fmt.Errorf("error")
	cases := []struct {
		k   int
		key Key
		err error
	}{
		{0000, Key{0, 0, 0, 0}, nil},
		{1234, Key{1, 2, 3, 4}, nil},
		{5678, Key{5, 6, 7, 8}, nil},
		{9999, Key{9, 9, 9, 9}, nil},
		{10000, nil, errorCase},
		{MaxInt, nil, errorCase},
		{-1, nil, errorCase},
		{MinInt, nil, errorCase},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%04d", c.k), func(t *testing.T) {
			key, err := NewKeyFromInt(c.k)
			if err != nil {
				if c.err == nil {
					t.Errorf("%s", err)
				}
				return
			}
			if c.err != nil {
				t.Errorf("NewKeyFromInt(%04d) == % X, want error", c.k, key)
				return
			}
			if !bytes.Equal(key, c.key) {
				t.Errorf("NewKeyFromInt(%04d) == % X, want % X", c.k, key, c.key)
			}

			k := key.Int()
			if k != c.k {
				t.Errorf("[% X].Int() == %04d, want %04d", key, k, c.k)
			}
		})
	}
}

func TestAllKeys(t *testing.T) {
	key := NewKey()
	n := key.Int()
	if n != 0 {
		t.Errorf("[% X].Int() == %04d, want %04d", key, n, 0)
	}
	for k := 0000; k <= 9999; k++ {
		n := key.Int()
		if n != k {
			t.Errorf("[% X].Int() == %04d, want %04d", key, n, k)
		}
		key2, err := NewKeyFromInt(k)
		if err != nil {
			t.Errorf("%s", err)
		}
		n = key2.Int()
		if n != k {
			t.Errorf("NewKeyFromInt(%04d).toInt() == %04d", k, n)
		}
		ok := key.Next()
		if !ok && k != 9999 {
			t.Errorf("[% X].Next() overflowed at %04d", key, k)
		}
	}
	if key.Int() != 0 {
		t.Errorf("key.Next() did not overflow at 9999")
	}
}

const benchmarkKey = 5622

func BenchmarkUnlockWithKey(b *testing.B) {
	p, err := Read(path.Join(testDataDir, benchmarkPuzzle))
	if err != nil {
		b.Errorf("%s", err)
		return
	}
	orig := p.solution
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.UnlockWithKey(benchmarkKey)
		p.solution = orig
		p.Scrambled = true
	}

}

func BenchmarkUnlock(b *testing.B) {
	p, err := Read(path.Join(testDataDir, benchmarkPuzzle))
	if err != nil {
		b.Errorf("%s", err)
		return
	}
	orig := p.solution
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Unlock()
		p.solution = orig
		p.Scrambled = true
	}

}
