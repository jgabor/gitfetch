package theme

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jgabor/gitfetch/internal/decay"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	BarWidth   = 15
	BarFilled  = "█"
	BarEmpty   = "░"
	EmptyColor = "240"
)

// TierColor returns the base lipgloss color for a decay tier.
func TierColor(tier decay.Tier) lipgloss.Color {
	return lipgloss.Color(tier.Color())
}

// ansiToRGB maps ANSI 256-color codes used by decay.Tier to RGB values.
func ansiToRGB(code string) colorful.Color {
	switch code {
	case "0":
		return colorful.Color{R: 0, G: 0, B: 0}
	case "1":
		return colorful.Color{R: 0.804, G: 0, B: 0} // #CD0000
	case "2":
		return colorful.Color{R: 0, G: 0.804, B: 0} // #00CD00
	case "3":
		return colorful.Color{R: 0.804, G: 0.804, B: 0} // #CDCD00
	case "208":
		return colorful.Color{R: 1, G: 0.529, B: 0} // #FF8700
	default:
		return colorful.Color{R: 0.5, G: 0.5, B: 0.5}
	}
}

// tierGradient returns the start and end colors for a tier's gradient.
func tierGradient(tier decay.Tier) (colorful.Color, colorful.Color) {
	base := ansiToRGB(tier.Color())
	end := colorful.Color{
		R: base.R * 0.4,
		G: base.G * 0.4,
		B: base.B * 0.4,
	}
	return base, end
}

// GradientBar returns a lipgloss-styled gradient bar for the given tier and
// progress percentage (0.0–1.0). Values outside [0,1] are clamped.
func GradientBar(tier decay.Tier, pct float64) string {
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

	start, end := tierGradient(tier)
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(EmptyColor))

	var b strings.Builder
	for i := 0; i < BarWidth; i++ {
		if i < filled {
			t := 0.0
			if BarWidth > 1 {
				t = float64(i) / float64(BarWidth-1)
			}
			c := start.BlendLab(end, t).Hex()
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(c))
			b.WriteString(style.Render(BarFilled))
		} else {
			b.WriteString(emptyStyle.Render(BarEmpty))
		}
	}
	return b.String()
}
