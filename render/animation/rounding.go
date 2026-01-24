package animation

import (
	"math"
)

type Rounding interface {
	Apply(v float64) float64
}

type Round struct{}

func (r Round) Apply(v float64) float64 {
	return math.Round(v)
}

type RoundFloor struct{}

func (rf RoundFloor) Apply(v float64) float64 {
	return math.Floor(v)
}

type RoundCeil struct{}

func (rc RoundCeil) Apply(v float64) float64 {
	return math.Ceil(v)
}

type RoundNone struct{}

func (rn RoundNone) Apply(v float64) float64 {
	return v
}

var DefaultRounding = Round{}
