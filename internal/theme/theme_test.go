package theme

import (
	"fmt"
	"strings"
	"testing"
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
			got := GradientBar(c.age, 0.5)
			if got == "" {
				t.Fatal("GradientBar returned empty string")
			}
			if !strings.Contains(got, BarFilled) {
				t.Error("expected filled characters in bar")
			}
		})

		t.Run(c.name+"_fail_negative_clamped", func(t *testing.T) {
			got := GradientBar(c.age, -0.5)
			want := GradientBar(c.age, 0)
			if got != want {
				t.Errorf("negative pct not clamped to 0")
			}
		})
	}
}

func TestGradientBar_EdgeCases(t *testing.T) {
	t.Run("zero_percent_no_filled", func(t *testing.T) {
		got := GradientBar(10, 0)
		if strings.Contains(got, BarFilled) {
			t.Error("0% bar should not contain filled characters")
		}
	})

	t.Run("hundred_percent_no_empty", func(t *testing.T) {
		got := GradientBar(10, 1)
		if strings.Contains(got, BarEmpty) {
			t.Error("100% bar should not contain empty characters")
		}
	})

	t.Run("negative_clamped", func(t *testing.T) {
		got := GradientBar(60, -1)
		want := GradientBar(60, 0)
		if got != want {
			t.Error("pct < 0 should clamp to 0")
		}
	})

	t.Run("over_one_clamped", func(t *testing.T) {
		got := GradientBar(120, 1.5)
		want := GradientBar(120, 1)
		if got != want {
			t.Error("pct > 1 should clamp to 1")
		}
	})
}

func TestGradientBar_DifferentAges(t *testing.T) {
	fresh := GradientBar(15, 0.5)
	stale := GradientBar(60, 0.5)
	decayed := GradientBar(135, 0.5)
	dead := GradientBar(200, 0.5)

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
