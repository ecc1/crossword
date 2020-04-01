package acrosslite

import (
	"fmt"
)

type (
	LayoutScore float64
)

func (a LayoutScore) String() string {
	return fmt.Sprintf("%.3f", a)
}

func (a LayoutScore) IsBetterThan(b LayoutScore) bool {
	return a > b
}

// layoutScore calculates a "goodness" score for the current layout.
// Larger point size and wider columns are preferred; under-filled columns are penalized.
func (r *RenderContext) layoutScore() LayoutScore {
	p := (r.cluePoints - minCluePoints) / (maxCluePoints - minCluePoints)
	w := r.columnWidth / r.pageWidth
	underfill := (r.pageHeight - r.y) / (r.pageHeight - r.columnTop())
	if underfill > 0.95 {
		// No penalty for empty columns.
		underfill = 0
	}
	return LayoutScore(1.1*p + 1.2*w - 0.8*underfill)
}
