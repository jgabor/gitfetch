package theme

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jgabor/gitfetch/internal/decay"
)

func TestGradientBar_VariousAges(t *testing.T) {
	cases := []struct {
		name string
		age  int
	}{
		{"fresh", 15},
		{"stale", 60},
		{"decayed", 135},
		{"dead", 200},
	}

	for _, c := range cases {
		t.Run(c.name+"_pass", func(t *testing.T) {
			got := GradientBar(c.age, 0.5, BarWidth)
			if got == "" {
				t.Fatal("GradientBar returned empty string")
			}
			if !strings.Contains(got, BarFilled) {
				t.Error("expected filled characters in bar")
			}
		})

		t.Run(c.name+"_fail_negative_clamped", func(t *testing.T) {
			got := GradientBar(c.age, -0.5, BarWidth)
			want := GradientBar(c.age, 0, BarWidth)
			if got != want {
				t.Errorf("negative pct not clamped to 0")
			}
		})
	}
}

func TestGradientBar_EdgeCases(t *testing.T) {
	t.Run("zero_percent_no_filled", func(t *testing.T) {
		got := GradientBar(10, 0, BarWidth)
		if strings.Contains(got, BarFilled) {
			t.Error("0% bar should not contain filled characters")
		}
	})

	t.Run("hundred_percent_no_empty", func(t *testing.T) {
		got := GradientBar(10, 1, BarWidth)
		if strings.Contains(got, BarEmpty) {
			t.Error("100% bar should not contain empty characters")
		}
	})

	t.Run("negative_clamped", func(t *testing.T) {
		got := GradientBar(60, -1, BarWidth)
		want := GradientBar(60, 0, BarWidth)
		if got != want {
			t.Error("pct < 0 should clamp to 0")
		}
	})

	t.Run("over_one_clamped", func(t *testing.T) {
		got := GradientBar(120, 1.5, BarWidth)
		want := GradientBar(120, 1, BarWidth)
		if got != want {
			t.Error("pct > 1 should clamp to 1")
		}
	})
}

func TestGradientBar_DifferentAges(t *testing.T) {
	fresh := GradientBar(15, 0.5, BarWidth)
	stale := GradientBar(60, 0.5, BarWidth)
	decayed := GradientBar(135, 0.5, BarWidth)
	dead := GradientBar(200, 0.5, BarWidth)

	if fresh == stale || stale == decayed || decayed == dead {
		t.Error("different ages should produce visually distinct bars")
	}
}

func TestAgeColor(t *testing.T) {
	cases := []struct {
		age int
	}{
		{0},
		{15},
		{30},
		{60},
		{90},
		{135},
		{180},
		{365},
	}
	for _, c := range cases {
		got := fmt.Sprintf("%v", AgeColor(c.age))
		if got == "" || got == "<nil>" {
			t.Errorf("AgeColor(%d) returned empty or nil: %q", c.age, got)
		}
	}
}

func TestAgeColorNegative(t *testing.T) {
	got := AgeColor(-5)
	want := AgeColor(0)
	if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
		t.Error("AgeColor(-5) should equal AgeColor(0)")
	}
}

func TestPaletteAlignedToTierBoundaries(t *testing.T) {
	palette := buildSpectrumPalette()
	if len(palette) != decay.DecayedLimit+1 {
		t.Fatalf("palette has %d entries, want %d (DecayedLimit)", len(palette), decay.DecayedLimit)
	}
	boundaries := []struct {
		day   int
		label string
	}{
		{0, "fresh-start"},
		{decay.FreshLimit - 1, "fresh-end"},
		{decay.FreshLimit, "stale-start"},
		{decay.StaleLimit - 1, "stale-end"},
		{decay.StaleLimit, "decayed-start"},
		{decay.DecayedLimit - 1, "decayed-end"},
	}
	for _, b := range boundaries {
		if b.day < 0 || b.day >= len(palette) {
			t.Errorf("%s day %d out of palette range", b.label, b.day)
		}
	}
	freshStart := fmt.Sprintf("%v", AgeColor(0))
	decayedEnd := fmt.Sprintf("%v", AgeColor(decay.DecayedLimit - 1))
	if freshStart == decayedEnd {
		t.Error("first and last palette colors should differ")
	}
}

func TestAgeColorBeyondDecayedLimit(t *testing.T) {
	last := AgeColor(decay.DecayedLimit)
	beyond := AgeColor(decay.DecayedLimit + 100)
	if fmt.Sprintf("%v", last) != fmt.Sprintf("%v", beyond) {
		t.Error("days beyond DecayedLimit should clamp to last palette color")
	}
}

func TestGradientBar_WidthAdaptsToWideTerminal(t *testing.T) {
	got := GradientBar(30, 0.5, 60)
	barOnly := stripANSI(got)
	if len([]rune(barOnly)) != 60 {
		t.Errorf("bar width = %d, want 60 for wide terminal", len([]rune(barOnly)))
	}
}

func TestGradientBar_WidthAdaptsToNarrowTerminal(t *testing.T) {
	got := GradientBar(30, 0.5, 10)
	barOnly := stripANSI(got)
	if len([]rune(barOnly)) != 10 {
		t.Errorf("bar width = %d, want 10 for narrow terminal", len([]rune(barOnly)))
	}
}

func TestGradientBar_WidthZeroClampedToOne(t *testing.T) {
	got := GradientBar(30, 0.5, 0)
	barOnly := stripANSI(got)
	if len([]rune(barOnly)) != 1 {
		t.Errorf("bar width = %d, want 1 when width=0", len([]rune(barOnly)))
	}
}

func TestGradientBar_WidthNegativeClampedToOne(t *testing.T) {
	got := GradientBar(30, 0.5, -5)
	barOnly := stripANSI(got)
	if len([]rune(barOnly)) != 1 {
		t.Errorf("bar width = %d, want 1 when width=-5", len([]rune(barOnly)))
	}
}

func stripANSI(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++
			}
			continue
		}
		result = append(result, s[i])
		i++
	}
	return string(result)
}
