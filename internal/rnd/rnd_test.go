package rnd

import (
	"fmt"
	"testing"
)

func TestparseCourseJSON(t *testing.T) {
	crs := parseCourseJSON("kit_carson_small.json")
	fmt.Println(crs)

	if crs.ID != "kit_carson" {
		t.Error("Got wrong course ID")
	}

	if crs.Name != "Kit Carson" {
		t.Error("Got wrong course name")
	}

	if crs.Holes[0].ID != "1" {
		t.Error("Got wrong first hole name")
	}

	if crs.Holes[0].Pars[0].Tee != "reg" {
		t.Error("Got wrong first par tee")
	}

	if crs.Holes[0].Pars[0].Pin != "A" {
		t.Error("Got wrong first par pin")
	}

	if crs.Holes[0].Pars[0].Par != 3 {
		t.Error("Got wrong first par")
	}
}

func TestParseRoundRaw(t *testing.T) {

	round := ParseRoundRaw("round_raw.csv")

	fmt.Println(round)

	if round.CrsID == "" {
		fmt.Println("Empty string detected for CourseID")
		t.FailNow()
	}

	if round.CrsID != "sunset_park" {
		t.Errorf("Got wrong course ID, got %s", round.CrsID)
	}

	if round.ID != "2021-08-03-06-23-29_-_sunset_park" {
		t.Errorf("Got wrong round ID, got %s", round.ID)
	}

	if round.CrsName != "Sunset Park Las Vegas" {
		t.Errorf("Got wrong course Name, got %s", round.CrsName)
	}

	if round.Holes[0].HoleID != "1" {
		t.Errorf("Got wrong first hole ID, got %s", round.Holes[0].HoleID)
	}

	if round.Holes[0].HoleID != "reg" {
		t.Error("Got wrong first Tee ID")
	}

	if round.Holes[0].PinID != "X" {
		t.Error("Got wrong first Pin ID")
	}

	if round.Holes[0].Score != 3 {
		t.Error("Got wrong first score")
	}

	if round.Holes[0].Par != 3 {
		t.Error("Got wrong first par")
	}
}
