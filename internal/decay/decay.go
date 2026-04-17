package decay

import (
	"time"
)

type Tier int

const (
	Fresh Tier = iota
	Stale
	Decayed
	Dead
)

const (
	FreshLimit   = 30
	StaleLimit   = 90
	DecayedLimit = 180
)

func (t Tier) String() string {
	switch t {
	case Fresh:
		return "fresh"
	case Stale:
		return "stale"
	case Decayed:
		return "decayed"
	case Dead:
		return "dead"
	default:
		return "unknown"
	}
}

func (t Tier) Label() string {
	return t.String()
}

func (t Tier) Color() string {
	switch t {
	case Fresh:
		return "2"
	case Stale:
		return "3"
	case Decayed:
		return "208"
	case Dead:
		return "1"
	default:
		return "0"
	}
}

func AgeDays(t time.Time) int {
	if t.IsZero() {
		return 0
	}
	return int(time.Since(t).Hours() / 24)
}

func Classify(t time.Time) Tier {
	return ClassifyByDays(AgeDays(t))
}

func ClassifyByDays(days int) Tier {
	switch {
	case days < FreshLimit:
		return Fresh
	case days < StaleLimit:
		return Stale
	case days < DecayedLimit:
		return Decayed
	default:
		return Dead
	}
}

func TierProgress(days int) float64 {
	switch {
	case days < FreshLimit:
		return float64(days) / float64(FreshLimit)
	case days < StaleLimit:
		return float64(days-FreshLimit) / float64(StaleLimit-FreshLimit)
	case days < DecayedLimit:
		return float64(days-StaleLimit) / float64(DecayedLimit-StaleLimit)
	default:
		return 1.0
	}
}

func DaysUntilNext(days int) int {
	switch {
	case days < FreshLimit:
		return FreshLimit - days
	case days < StaleLimit:
		return StaleLimit - days
	case days < DecayedLimit:
		return DecayedLimit - days
	default:
		return 0
	}
}
