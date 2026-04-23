package theme

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/jgabor/gitfetch/internal/decay"
)

const (
	BarWidth   = 30
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
)

// spectrumPalette maps each day (0..DecayedLimit) to a color aligned with
// tier boundaries: green through fresh, yellow→orange through stale,
// orange→red through decayed.
var spectrumPalette = buildSpectrumPalette()

func buildSpectrumPalette() []color.Color {
	type segment struct {
		steps    int
		from, to color.Color
	}
	segs := []segment{
		{decay.FreshLimit, colorGreen, colorYellow},
		{decay.StaleLimit - decay.FreshLimit, colorYellow, colorOrange},
		{decay.DecayedLimit - decay.StaleLimit, colorOrange, colorRed},
	}
	var palette []color.Color
	for _, s := range segs {
		blend := lipgloss.Blend1D(s.steps+1, s.from, s.to)
		if palette != nil {
			blend = blend[1:]
		}
		palette = append(palette, blend...)
	}
	return palette
}

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
// The gradient maps each filled character to a color from the tier-aligned
// spectrum palette. Days beyond DecayedLimit are clamped so the bar saturates
// at the maximum decay color (red).
func GradientBar(days int, pct float64, width int) string {
	if width < 1 {
		width = 1
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	filled := int(math.Round(float64(width) * pct))
	if filled > width {
		filled = width
	}

	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(EmptyColor))

	var b strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			t := 0.0
			if width > 1 {
				t = float64(i) / float64(width-1)
			}
			age := int(t * float64(days))
			if age >= len(spectrumPalette) {
				age = len(spectrumPalette) - 1
			}
			c := AgeColor(age)
			style := lipgloss.NewStyle().Foreground(c)
			b.WriteString(style.Render(BarFilled))
		} else {
			b.WriteString(emptyStyle.Render(BarEmpty))
		}
	}
	return b.String()
}
