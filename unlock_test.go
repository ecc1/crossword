package acrosslite

import (
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
