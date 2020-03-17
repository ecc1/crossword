package acrosslite

import (
	"io/ioutil"
	"path"
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

type (
	Square struct {
		X int
		Y int
	}
	Squares []Square
)

func TestReadPuzzle(t *testing.T) {
	cases := []struct {
		file    string
		puz     Puzzle
		circled Squares
	}{
		{
			"Apr2510.puz",
			Puzzle{
				Author:    "Elizabeth C. Gorski / Will Shortz",
				Copyright: "© 2010, The New York Times",
				Title:     "NY Times, Sun, Apr 25, 2010 MONUMENTAL ACHIEVEMENT (See Notepad)",
				Notepad:   "When this puzzle is done, the seven circled letters can be arranged to spell a common word, which is missing from seven of the clues, as indicated by [ ]. Connect the seven letters in order with a line and you will get an outline of the object that the word names\r\n",
				Width:     21,
				Height:    21,
				AcrossClues: IndexedStrings{
					1:   "Tubs",
					6:   "Dead",
					11:  "Large amount",
					15:  "Imported cheese",
					19:  "Tribe of Israel",
					20:  "Resident of a country that's 97% mountains and desert",
					21:  "Sailor's direction",
					22:  "\"Here I ___ Worship\" (contemporary hymn)",
					23:  "[ ]",
					27:  "Fling",
					28:  "English connections",
					29:  "\"Le Déjeuner des Canotiers,\" e.g.",
					30:  "You may get a charge out of it",
					31:  "Gwen who sang \"Don't Speak,\" 1996",
					33:  "Top of a mountain?",
					35:  "Saintly glows",
					37:  "[ ]",
					41:  "Leaving for",
					44:  "\"Go on!\"",
					45:  "\"A pity\"",
					46:  "Charles, for one",
					47:  "Very friendly (with)",
					49:  "Start of a famous J.F.K. quote",
					52:  "Price part: Abbr.",
					55:  "[ ]",
					58:  "Pizza orders",
					59:  "Glossy black birds",
					60:  "New York City transport from the Bronx to Coney Island",
					61:  "Throat soother",
					63:  "Like clogs",
					65:  "After, in Avignon",
					66:  "Paris attraction that features a [ ]",
					69:  "Passes over",
					70:  "Football shoes",
					72:  "Nervousness",
					73:  "Low clouds",
					75:  "Fannie ___ (some investments)",
					76:  "Prenatal procedures, informally",
					78:  "[ ]",
					80:  "Coast Guard rank: Abbr.",
					81:  "Snow fall",
					82:  "Run ___ of",
					84:  "Willy who wrote \"The Conquest of Space\"",
					85:  "Whites or colors, e.g.",
					86:  "NASA's ___ Research Center",
					87:  "Trumpet",
					89:  "[ ] that was the creation of an architect born 4/26/1917",
					97:  "Humdingers",
					98:  "Atomic centers",
					99:  "Mozart's birthplace",
					103: "Network that airs \"WWE Raw\"",
					104: "Breakdown of social norms",
					106: "Naval officer: Abbr.",
					108: "Bop",
					109: "[ ]",
					114: "O'Neill's \"Desire Under the ___\"",
					115: "\"___ Death\" (Grieg movement)",
					116: "Flat storage place",
					117: "Headless Horseman, e.g.",
					118: "Way: Abbr.",
					119: "Larry who played Tony in \"West Side Story\"",
					120: "Compost units",
					121: "Professional grps.",
				},
				DownClues: IndexedStrings{
					1:   "Almanac tidbits",
					2:   "\"Give it ___\"",
					3:   "\"___ Foolish Things\" (1936 hit)",
					4:   "Deems worthy",
					5:   "Canadian-born hockey great",
					6:   "Walter of \"Star Trek\"",
					7:   "\"Diary of ___ Housewife\"",
					8:   "Crash sites?",
					9:   "Prefix with sex",
					10:  "Cookie holder",
					11:  "Seattle's ___ Field",
					12:  "Like some cell growth",
					13:  "Part of a Virgin Atlantic fleet",
					14:  "Prefix with monde",
					15:  "\"Let's ___!\"",
					16:  "Composer Shostakovich",
					17:  "Like Berg's \"Wozzeck\"",
					18:  "Williams of TV",
					24:  "Smallville girl",
					25:  "Sudoku feature",
					26:  "Genesis landing site",
					32:  "\"I love,\" in Latin",
					33:  "Tizzy",
					34:  "\"Krazy\" one",
					36:  "Financial inst. that bought PaineWebber in 2000",
					38:  "Upper hand",
					39:  "\"I'm impressed!\"",
					40:  "At ___ for words",
					41:  "Suffix with contradict",
					42:  "Nutritional regimen",
					43:  "Parts of some Mediterranean orchards",
					47:  "French pronoun",
					48:  "Exists no more",
					49:  "High: Lat.",
					50:  "It doesn't hold water",
					51:  "1980s Chrysler debut",
					52:  "April first?",
					53:  "Double-crosser",
					54:  "Payroll stub IDs",
					56:  "Fields",
					57:  "History",
					58:  "Covered walkways",
					59:  "Joltin' Joe",
					61:  "\"Thin Ice\" star Sonja",
					62:  "Bars from the refrigerator",
					64:  "\"___, is it I?\"",
					65:  "Tip-top",
					67:  "Pinup boy",
					68:  "\"___ Wood sawed wood\" (start of a tongue twister)",
					71:  "Light lunch",
					74:  "Bygone daily MTV series, informally",
					77:  "Clapped and shouted, e.g.",
					78:  "\"___ fan tutte\"",
					79:  "Ophthalmologist's study",
					81:  "Anatomical cavities",
					82:  "Both: Prefix",
					83:  "Tina of \"30 Rock\"",
					85:  "Baton Rouge sch.",
					86:  "\"Wheel of Fortune\" purchase",
					87:  "Wanna-___ (imitators)",
					88:  "They're nuts",
					89:  "Sitting areas, slangily?",
					90:  "How rain forests grow",
					91:  "Bells and whistles, maybe",
					92:  "Kind of romance",
					93:  "Least friendly",
					94:  "Valley",
					95:  "House keepers",
					96:  "Knitting loop",
					100: "Some have forks",
					101: "How some people solve crosswords",
					102: "Singer/actress Karen of Broadway's \"Nine\"",
					105: "Neighbor of Sask.",
					106: "Mrs. Dithers of \"Blondie\"",
					107: "Run before Q",
					110: "Ballpark fig.",
					111: "Brown, e.g.: Abbr.",
					112: "Chemical suffix",
					113: "Spanish Mrs.",
				},
			},
			Squares{{8, 0}, {8, 1}, {0, 12}, {20, 13}, {8, 17}, {11, 19}, {12, 19}},
		},
		{
			"Mar1420.puz",
			Puzzle{
				Author:    "Peter Wentz / Will Shortz",
				Copyright: "© 2020, The New York Times",
				Title:     "NY Times, Saturday, March 14, 2020 ",
				Width:     15,
				Height:    15,
				AcrossClues: IndexedStrings{
					1:  "Openness",
					7:  "Launch",
					13: "State capital whose name is pronounced as one syllable (not two, as many think)",
					14: "Pamper",
					15: "What a good tip can lead to",
					16: "Unnamed women",
					17: "Activity for kids out for kicks?",
					19: "Dental hygienist's order",
					20: "___ justice",
					21: "Tastes, say",
					23: "Chain named phonetically after its founders",
					25: "Refuse to go there!",
					26: "Green org.",
					30: "So-called \"good cholesterol\"",
					31: "\"Ah, all right\"",
					33: "Participant in a 1990s civil war",
					34: "Thai neighbor",
					35: "Final part of a track race",
					37: "It comes three after pi",
					38: "Member of an old Western empire",
					40: "Popular photo-sharing site",
					41: "Waiting room features",
					42: "Calls on",
					43: "Tea company owned by Unilever",
					44: "George W. Bush or George H. W. Bush",
					46: "Handout at check-in",
					49: "Rewards for good behavior, maybe",
					50: "Lumberjack",
					53: "Guy who's easily dismissed",
					55: "It's office-bound",
					57: "\"Amscray!\"",
					59: "\"Sounds 'bout right\"",
					60: "N.L. Central player",
					61: "Bouncer's confiscation",
					62: "Costing a great deal, informally",
				},
				AcrossAnswers: IndexedStrings{
					1:  "CANDOR",
					7:  "BOOTUP",
					13: "PIERRE",
					14: "CATERTO",
					15: "ARREST",
					16: "JANEDOES",
					17: "JVSOCCER",
					19: "RINSE",
					20: "DOES",
					21: "HASASIP",
					23: "ARBYS",
					25: "DUMP",
					26: "USGA",
					30: "HDL",
					31: "OHISEE",
					33: "SERB",
					34: "LAO",
					35: "BELLLAP",
					37: "TAU",
					38: "INCA",
					40: "FLICKR",
					41: "TVS",
					42: "ASKS",
					43: "TAZO",
					44: "YALIE",
					46: "KEYCARD",
					49: "PETS",
					50: "AXMAN",
					53: "MRNOBODY",
					55: "STENOPAD",
					57: "BUGOFF",
					59: "IRECKON",
					60: "BREWER",
					61: "FAKEID",
					62: "SPENDY",
				},
				DownClues: IndexedStrings{
					1:  "Reconciler, for short",
					2:  "Prized footwear introduced in 1984",
					3:  "Chronic pain remedy",
					4:  "Formal",
					5:  "Around there",
					6:  "Heave",
					7:  "Force onto the black market, say",
					8:  "\"S.N.L.\" castmate of Shannon and Gasteyer",
					9:  "Complex figure?",
					10: "Classic film with a game theme",
					11: "Neighbors of the Navajo",
					12: "Present",
					14: "Carnival bagful",
					16: "Informal name for a reptile that can seemingly run on water",
					18: "1990 Robin Williams title role",
					20: "Mexico's national flower",
					22: "Make a delivery",
					24: "Blubber",
					27: "\"Quit horsing around!\"",
					28: "Not needing a pump",
					29: "Causes for censuring, maybe",
					32: "Glad competitor",
					36: "Wrench with power",
					39: "With disapproval or distrust",
					45: "Roughly 251,655 miles, for Earth's moon",
					47: "Ramen topping",
					48: "\"Independents Day\" author Lou",
					50: "\"That's rich!\"",
					51: "Bonus, in ad lingo",
					52: "Compliant",
					54: "Pat on the back",
					56: "Peeved",
					58: "Get burned",
				},
				DownAnswers: IndexedStrings{
					1:  "CPA",
					2:  "AIRJORDANS",
					3:  "NERVEBLOCK",
					4:  "DRESSY",
					5:  "ORSO",
					6:  "RETCH",
					7:  "BAN",
					8:  "OTERI",
					9:  "OEDIPUS",
					10: "TRON",
					11: "UTES",
					12: "POSE",
					14: "CARAMELCORN",
					16: "JESUSLIZARD",
					18: "CADILLACMAN",
					20: "DAHLIA",
					22: "SPEAK",
					24: "SOB",
					27: "SETTLEDOWN",
					28: "GRAVITYFED",
					29: "ABUSES",
					32: "HEFTY",
					36: "PRY",
					39: "ASKANCE",
					45: "APOGEE",
					47: "ENOKI",
					48: "DOBBS",
					50: "ASIF",
					51: "XTRA",
					52: "MEEK",
					54: "BURP",
					56: "POD",
					58: "FRY",
				},
			},
			Squares{},
		},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			p, err := Read(path.Join(testDataDir, c.file))
			if err != nil {
				t.Errorf("%s", err)
				return
			}
			if p.Author != c.puz.Author {
				t.Errorf("Author == %q, want %q", p.Author, c.puz.Author)
			}
			if p.Copyright != c.puz.Copyright {
				t.Errorf("Copyright == %q, want %q", p.Copyright, c.puz.Copyright)
			}
			if p.Title != c.puz.Title {
				t.Errorf("Title == %q, want %q", p.Title, c.puz.Title)
			}
			if p.Notepad != c.puz.Notepad {
				t.Errorf("Notepad == %q, want %q", p.Notepad, c.puz.Notepad)
			}
			if p.Width != c.puz.Width {
				t.Errorf("Width == %d, want %d", p.Width, c.puz.Width)
			}
			if p.Height != c.puz.Height {
				t.Errorf("Height == %d, want %d", p.Height, c.puz.Height)
			}
			checkCircles(t, p, c.circled)
			checkMap(t, "Across clue", p.AcrossClues, c.puz.AcrossClues)
			checkMap(t, "Across answer", p.AcrossAnswers, c.puz.AcrossAnswers)
			checkMap(t, "Down clue", p.DownClues, c.puz.DownClues)
			checkMap(t, "Down answer", p.DownAnswers, c.puz.DownAnswers)
		})
	}
}

func inSquares(x int, y int, circled Squares) bool {
	for _, sq := range circled {
		if sq.X == x && sq.Y == y {
			return true
		}
	}
	return false
}

func checkCircles(t *testing.T, p *Puzzle, circled Squares) {
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			if inSquares(x, y, circled) {
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

func checkMap(t *testing.T, kind string, got IndexedStrings, want IndexedStrings) {
	for n, w := range want {
		g := got[n]
		if w != g {
			t.Errorf("%s %d: want %q, got %q", kind, n, w, g)
		}
	}
	for n, g := range got {
		w := want[n]
		if w != "" {
			continue
		}
		t.Errorf("%s %d: got %q (unexpected)", kind, n, g)
	}
}
