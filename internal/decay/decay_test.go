package decay

import (
	"testing"
	"time"
)

func TestClassifyFresh(t *testing.T) {
	date := time.Now().UTC().Add(-10 * 24 * time.Hour)
	if tier := Classify(date); tier != Fresh {
		t.Errorf("10 days ago = %s, want fresh", tier)
	}
}

func TestClassifyStale(t *testing.T) {
	date := time.Now().UTC().Add(-60 * 24 * time.Hour)
	if tier := Classify(date); tier != Stale {
		t.Errorf("60 days ago = %s, want stale", tier)
	}
}

func TestClassifyDecayed(t *testing.T) {
	date := time.Now().UTC().Add(-120 * 24 * time.Hour)
	if tier := Classify(date); tier != Decayed {
		t.Errorf("120 days ago = %s, want decayed", tier)
	}
}

func TestClassifyDead(t *testing.T) {
	date := time.Now().UTC().Add(-200 * 24 * time.Hour)
	if tier := Classify(date); tier != Dead {
		t.Errorf("200 days ago = %s, want dead", tier)
	}
}

func TestBoundaryFreshStale29(t *testing.T) {
	if tier := ClassifyByDays(29); tier != Fresh {
		t.Errorf("29 days = %s, want fresh", tier)
	}
}

func TestBoundaryFreshStale30(t *testing.T) {
	if tier := ClassifyByDays(30); tier != Stale {
		t.Errorf("30 days = %s, want stale", tier)
	}
}

func TestBoundaryFreshStale31(t *testing.T) {
	if tier := ClassifyByDays(31); tier != Stale {
		t.Errorf("31 days = %s, want stale", tier)
	}
}

func TestBoundaryStaleDecayed89(t *testing.T) {
	if tier := ClassifyByDays(89); tier != Stale {
		t.Errorf("89 days = %s, want stale", tier)
	}
}

func TestBoundaryStaleDecayed90(t *testing.T) {
	if tier := ClassifyByDays(90); tier != Decayed {
		t.Errorf("90 days = %s, want decayed", tier)
	}
}

func TestBoundaryStaleDecayed91(t *testing.T) {
	if tier := ClassifyByDays(91); tier != Decayed {
		t.Errorf("91 days = %s, want decayed", tier)
	}
}

func TestBoundaryDecayedDead179(t *testing.T) {
	if tier := ClassifyByDays(179); tier != Decayed {
		t.Errorf("179 days = %s, want decayed", tier)
	}
}

func TestBoundaryDecayedDead180(t *testing.T) {
	if tier := ClassifyByDays(180); tier != Dead {
		t.Errorf("180 days = %s, want dead", tier)
	}
}

func TestBoundaryDecayedDead181(t *testing.T) {
	if tier := ClassifyByDays(181); tier != Dead {
		t.Errorf("181 days = %s, want dead", tier)
	}
}

func TestClassifyByDaysZero(t *testing.T) {
	if tier := ClassifyByDays(0); tier != Fresh {
		t.Errorf("0 days = %s, want fresh", tier)
	}
}

func TestClassifyByDaysLarge(t *testing.T) {
	if tier := ClassifyByDays(365); tier != Dead {
		t.Errorf("365 days = %s, want dead", tier)
	}
}

func TestTierString(t *testing.T) {
	tests := []struct {
		tier    Tier
		want    string
		wantLen int
	}{
		{Fresh, "fresh", 5},
		{Stale, "stale", 5},
		{Decayed, "decayed", 7},
		{Dead, "dead", 4},
	}
	for _, tt := range tests {
		if got := tt.tier.String(); got != tt.want {
			t.Errorf("Tier(%d).String() = %q, want %q", tt.tier, got, tt.want)
		}
	}
}

func TestTierColor(t *testing.T) {
	colors := map[Tier]string{
		Fresh:   "2",
		Stale:   "3",
		Decayed: "208",
		Dead:    "1",
	}
	for tier, want := range colors {
		if got := tier.Color(); got != want {
			t.Errorf("Tier(%d).Color() = %q, want %q", tier, got, want)
		}
	}
}

func TestAgeDaysZeroTime(t *testing.T) {
	if days := AgeDays(time.Time{}); days != 0 {
		t.Errorf("AgeDays(zero) = %d, want 0", days)
	}
}

func TestAgeDaysNow(t *testing.T) {
	if days := AgeDays(time.Now().UTC()); days != 0 {
		t.Errorf("AgeDays(now) = %d, want 0", days)
	}
}

func TestTierProgressFresh(t *testing.T) {
	p := TierProgress(0)
	if p != 0.0 {
		t.Errorf("TierProgress(0) = %f, want 0.0", p)
	}
	p = TierProgress(15)
	if p < 0.49 || p > 0.51 {
		t.Errorf("TierProgress(15) = %f, want ~0.5", p)
	}
}

func TestTierProgressStale(t *testing.T) {
	p := TierProgress(30)
	if p != 0.0 {
		t.Errorf("TierProgress(30) = %f, want 0.0", p)
	}
	p = TierProgress(60)
	if p < 0.49 || p > 0.51 {
		t.Errorf("TierProgress(60) = %f, want ~0.5", p)
	}
}

func TestTierProgressDecayed(t *testing.T) {
	p := TierProgress(90)
	if p != 0.0 {
		t.Errorf("TierProgress(90) = %f, want 0.0", p)
	}
	p = TierProgress(135)
	if p < 0.49 || p > 0.51 {
		t.Errorf("TierProgress(135) = %f, want ~0.5", p)
	}
}

func TestTierProgressDead(t *testing.T) {
	p := TierProgress(180)
	if p != 1.0 {
		t.Errorf("TierProgress(180) = %f, want 1.0", p)
	}
	p = TierProgress(365)
	if p != 1.0 {
		t.Errorf("TierProgress(365) = %f, want 1.0", p)
	}
}
