package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

type ThrowRow struct {
	Date     string
	Course   string
	Round    string
	Hole     string
	Shot     string
	Disc     string
	Lat1     string
	Lon1     string
	Lat2     string
	Lon2     string
	LatPin   string
	LonPin   string
	Dist     string
	DistPin  string
	ResThrow string
}
type AllThrows []ThrowRow

func (t ThrowRow) asStrings() []string {
	return []string{
		t.Date,
		t.Course,
		t.Round,
		t.Hole,
		t.Shot,
		t.Disc,
		t.Lat1,
		t.Lon1,
		t.Lat2,
		t.Lon2,
		t.LatPin,
		t.LonPin,
		t.Dist,
		t.DistPin,
		t.ResThrow,
	}
}

func (a AllThrows) GetHeader() ThrowRow {
	return ThrowRow{
		Date:     "Date",
		Course:   "CourseID",
		Round:    "RoundID",
		Hole:     "HoleID",
		Shot:     "Shot#",
		Disc:     "Disc",
		Lat1:     "Lat1",
		Lon1:     "Lon1",
		Lat2:     "Lat2",
		Lon2:     "Lon2",
		LatPin:   "LatPin",
		LonPin:   "LonPin",
		Dist:     "Dist",
		DistPin:  "DistPin",
		ResThrow: "ResThrow",
	}
}

func GetRoundThrows(rd rnd.Round) AllThrows {

	var AT AllThrows

	date := rd.ID[:10]
	courseID := rd.CourseID
	roundID := rd.ID
	course := rd.Course

	cur_hole := ""
	shot_num := 0
	for i, r := range rd.Data {

		if r.Disc == "BASKET" {
			continue
		}

		nr := rd.Data[i+1]

		h := course.GetHole(r.HoleID)
		p := h.GetPin(r.PinID)

		if r.HoleID != cur_hole {
			cur_hole = r.HoleID
			shot_num = 1
		} else {
			shot_num++
		}

		d := math.Round(3.281 * rnd.Dist(r.Loc(), nr.Loc()))

		dpin := math.Round(3.281 * rnd.Dist(r.Loc(), p.Loc))

		res := "THROW"
		if shot_num == 1 && nr.Disc == "BASKET" {
			res = "ACE"
		} else if shot_num == 1 {
			res = "DRIVE"
		} else if nr.Disc == "BASKET" {
			res = "MAKE"
		}

		AT = append(AT, ThrowRow{
			Date:     date,
			Course:   courseID,
			Round:    roundID,
			Hole:     r.HoleID,
			Shot:     strconv.Itoa(shot_num),
			Disc:     r.Disc,
			Lat1:     fmt.Sprintf("%f", r.Lat),
			Lon1:     fmt.Sprintf("%f", r.Lon),
			Lat2:     fmt.Sprintf("%f", nr.Lat),
			Lon2:     fmt.Sprintf("%f", nr.Lon),
			LatPin:   fmt.Sprintf("%f", p.Loc[0]),
			LonPin:   fmt.Sprintf("%f", p.Loc[1]),
			Dist:     fmt.Sprintf("%f", d),
			DistPin:  fmt.Sprintf("%f", dpin),
			ResThrow: res,
		})
	}

	return AT
}

func MakeAllThrowsCSV(rnds []rnd.Round) {

	var AT AllThrows
	for _, rd := range rnds {
		AT = append(AT, GetRoundThrows(rd)...)
	}

	homedir, _ := os.UserHomeDir()
	at_csv := filepath.Join(homedir, ".discgolf", "stats", "all_throws.csv")

	f, err := os.Create(at_csv)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write(AT.GetHeader().asStrings())

	for _, t := range AT {
		w.Write(t.asStrings())
	}

}
