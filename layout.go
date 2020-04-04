package crossword

import (
	"fmt"
)

type (
	LayoutScore struct {
		pointSize   float64
		columnWidth float64
		columnUsage float64
	}
)

func (a LayoutScore) String() string {
	return fmt.Sprintf("%.3f [%.3f %.3f %.3f]", a.Scalar(),
		a.pointSize, a.columnWidth, a.columnUsage)
}

func (a LayoutScore) IsBetterThan(b LayoutScore) bool {
	return a.Scalar() > b.Scalar()
}

// Scalar calculates a single "goodness" score from the LayoutScore components.
func (a LayoutScore) Scalar() float64 {
	return a.pointSize + a.columnWidth + 0.5*a.columnUsage
}

// layoutScore returns the components used to calculate the "goodness" of the layout.
// Each component is scaled to the interval [0, 1].
func (r *RenderContext) layoutScore() LayoutScore {
	return LayoutScore{
		pointSize:   (r.cluePoints - minCluePoints) / (maxCluePoints - minCluePoints),
		columnWidth: r.columnWidth / r.pageWidth,
		columnUsage: 1 / float64(r.currentColumn),
	}
}
