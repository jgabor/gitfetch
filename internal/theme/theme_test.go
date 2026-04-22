package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jgabor/gitfetch/internal/decay"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	m.Run()
}

func TestTierColor(t *testing.T) {
	tests := []struct {
		tier decay.Tier
		want string
	}{
		{decay.Fresh, "2"},
		{decay.Stale, "3"},
		{decay.Decayed, "208"},
		{decay.Dead, "1"},
	}
	for _, tt := range tests {
		t.Run(tt.tier.String(), func(t *testing.T) {
			got := TierColor(tt.tier)
			if string(got) != tt.want {
				t.Errorf("TierColor(%s) = %q, want %q", tt.tier.String(), got, tt.want)
			}
		})
	}
}

func TestGradientBar_EachTier(t *testing.T) {
	cases := []struct {
		name string
		tier decay.Tier
	}{
		{"fresh", decay.Fresh},
		{"stale", decay.Stale},
		{"decayed", decay.Decayed},
		{"dead", decay.Dead},
	}

	for _, c := range cases {
		t.Run(c.name+"_pass", func(t *testing.T) {
			got := GradientBar(c.tier, 0.5)
			if got == "" {
				t.Fatal("GradientBar returned empty string")
			}
			if !strings.Contains(got, BarFilled) {
				t.Error("expected filled characters in bar")
			}
		})

		t.Run(c.name+"_fail_negative_clamped", func(t *testing.T) {
			got := GradientBar(c.tier, -0.5)
			want := GradientBar(c.tier, 0)
			if got != want {
				t.Errorf("negative pct not clamped to 0")
			}
		})
	}
}

func TestGradientBar_EdgeCases(t *testing.T) {
	t.Run("zero_percent_no_filled", func(t *testing.T) {
		got := GradientBar(decay.Fresh, 0)
		if strings.Contains(got, BarFilled) {
			t.Error("0% bar should not contain filled characters")
		}
	})

	t.Run("hundred_percent_no_empty", func(t *testing.T) {
		got := GradientBar(decay.Fresh, 1)
		if strings.Contains(got, BarEmpty) {
			t.Error("100% bar should not contain empty characters")
		}
	})

	t.Run("negative_clamped", func(t *testing.T) {
		got := GradientBar(decay.Stale, -1)
		want := GradientBar(decay.Stale, 0)
		if got != want {
			t.Error("pct < 0 should clamp to 0")
		}
	})

	t.Run("over_one_clamped", func(t *testing.T) {
		got := GradientBar(decay.Decayed, 1.5)
		want := GradientBar(decay.Decayed, 1)
		if got != want {
			t.Error("pct > 1 should clamp to 1")
		}
	})
}

func TestGradientBar_DifferentTiers(t *testing.T) {
	fresh := GradientBar(decay.Fresh, 0.5)
	stale := GradientBar(decay.Stale, 0.5)
	decayed := GradientBar(decay.Decayed, 0.5)
	dead := GradientBar(decay.Dead, 0.5)

	if fresh == stale || stale == decayed || decayed == dead {
		t.Error("different tiers should produce visually distinct bars")
	}
}
