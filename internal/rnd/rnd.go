package rnd

import (
	"math"
)

type Loc []float64

func dist(l1 Loc, l2 Loc) float64 {
	R := 6378000.0

	φ1 := l1[0] * 3.14159 / 180.0
	φ2 := l2[0] * 3.14159 / 180.0
	Δφ := (l2[0] - l1[0]) * 3.14159 / 180.0
	Δλ := (l2[1] - l1[1]) * 3.14159 / 180.0

	a := math.Sin(Δφ/2.0)*math.Sin(Δφ/2.0) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2.0)*math.Sin(Δλ/2.0)
	c := 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))

	return R * c
}

type Pin struct {
	ID  string `json:"id"`
	Loc Loc    `json:"loc"`
}

type Tee struct {
	ID  string `json:"id"`
	Loc Loc    `json:"loc"`
}

type Par struct {
	Tee string `json:"tee"`
	Pin string `json:"pin"`
	Par int    `json:"par"`
}

type Hole struct {
	ID   string `json:"id"`
	Tees []Tee  `json:"tees"`
	Pins []Pin  `json:"pins"`
	Pars []Par  `json:"pars"`
}

type Course struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Loc   Loc    `json:"loc"`
	Holes []Hole `json:"holes"`
}

func (h Hole) Par(tID string, pID string) int {
	for _, p := range h.Pars {
		if p.Tee == tID && p.Pin == pID {
			return p.Par
		}
	}
	return 99
}

func (c Course) GetHole(hID string) Hole {
	for _, h := range c.Holes {
		if h.ID == hID {
			return h
		}
	}
	return Hole{}
}

func (h Hole) GetTee(tID string) Tee {
	for _, t := range h.Tees {
		if t.ID == tID {
			return t
		}
	}
	return Tee{}
}

func (h Hole) GetPin(pID string) Pin {
	for _, p := range h.Pins {
		if p.ID == pID {
			return p
		}
	}
	return Pin{}
}
