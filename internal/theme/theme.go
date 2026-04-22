package theme

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/jgabor/gitfetch/internal/decay"
)

const (
	BarWidth   = 15
	BarFilled  = "█"
	BarEmpty   = "░"
	EmptyColor = "240"
)

// Spectrum stops for the age-based color gradient.
var (
	colorGreen    = lipgloss.Color("#00cd00")
	colorYellow   = lipgloss.Color("#cdcd00")
	colorOrange   = lipgloss.Color("#ff8700")
	colorRed      = lipgloss.Color("#cd0000")
	colorDarkRed  = lipgloss.Color("#660000")
)

// spectrumPalette is a pre-generated 361-color palette covering the full
// decay spectrum from 0 to 360 days.
var spectrumPalette = lipgloss.Blend1D(361,
	colorGreen,
	colorYellow,
	colorOrange,
	colorRed,
	colorDarkRed,
)

// TierColor returns the base lipgloss color for a decay tier.
func TierColor(tier decay.Tier) color.Color {
	return lipgloss.Color(tier.Color())
}

// AgeColor returns the interpolated color for a given age in days.
func AgeColor(days int) color.Color {
	if days < 0 {
		days = 0
	}
	if days >= len(spectrumPalette) {
		days = len(spectrumPalette) - 1
	}
	return spectrumPalette[days]
}

// GradientBar returns a lipgloss-styled gradient bar for the given repo age and
// progress percentage (0.0–1.0). Values outside [0,1] are clamped.
// The gradient always starts at green (age 0) and progresses through the full
// age color spectrum up to the repo's actual age.
func GradientBar(days int, pct float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	filled := int(math.Round(float64(BarWidth) * pct))
	if filled > BarWidth {
		filled = BarWidth
	}

	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(EmptyColor))

	var b strings.Builder
	for i := 0; i < BarWidth; i++ {
		if i < filled {
			t := 0.0
			if BarWidth > 1 {
				t = float64(i) / float64(BarWidth-1)
			}
			age := int(t * float64(days))
			c := AgeColor(age)
			style := lipgloss.NewStyle().Foreground(c)
			b.WriteString(style.Render(BarFilled))
		} else {
			b.WriteString(emptyStyle.Render(BarEmpty))
		}
	}
	return b.String()
}
