package main

import (
	"encoding/csv"
	"fmt"
	"io"
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
	Shot     int
	Disc     string
	Lat1     float64
	Lon1     float64
	Lat2     float64
	Lon2     float64
	LatPin   float64
	LonPin   float64
	Dist     float64
	DistPin  float64
	ResThrow string
}
type AllThrows []ThrowRow

func (t ThrowRow) asStrings() []string {
	return []string{
		t.Date,
		t.Course,
		t.Round,
		t.Hole,
		strconv.Itoa(t.Shot),
		t.Disc,
		fmt.Sprintf("%f", t.Lat1),
		fmt.Sprintf("%f", t.Lon1),
		fmt.Sprintf("%f", t.Lat2),
		fmt.Sprintf("%f", t.Lon2),
		fmt.Sprintf("%f", t.LatPin),
		fmt.Sprintf("%f", t.LonPin),
		fmt.Sprintf("%g", t.Dist),
		fmt.Sprintf("%g", t.DistPin),
		t.ResThrow,
	}
}

func (a AllThrows) GetHeader() []string {
	return []string{
		"Date",
		"CourseID",
		"RoundID",
		"HoleID",
		"Shot#",
		"Disc",
		"Lat1",
		"Lon1",
		"Lat2",
		"Lon2",
		"LatPin",
		"LonPin",
		"Dist",
		"DistPin",
		"ResThrow",
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
			Shot:     shot_num,
			Disc:     r.Disc,
			Lat1:     r.Lat,
			Lon1:     r.Lon,
			Lat2:     nr.Lat,
			Lon2:     nr.Lon,
			LatPin:   p.Loc[0],
			LonPin:   p.Loc[1],
			Dist:     d,
			DistPin:  dpin,
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

	w.Write(AT.GetHeader())

	for _, t := range AT {
		w.Write(t.asStrings())
	}

}

func ParseAllThrowsCSV() AllThrows {
	homedir, _ := os.UserHomeDir()
	arCSV := filepath.Join(homedir, ".discgolf", "stats", "all_throws.csv")

	f, _ := os.Open(arCSV)
	defer f.Close()
	var AT AllThrows
	r := csv.NewReader(f)
	r.Read() // burn header
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}

		shot_num, _ := strconv.Atoi(line[4])
		lat1, _ := strconv.ParseFloat(line[6], 64)
		lon1, _ := strconv.ParseFloat(line[7], 64)
		lat2, _ := strconv.ParseFloat(line[8], 64)
		lon2, _ := strconv.ParseFloat(line[9], 64)
		latp, _ := strconv.ParseFloat(line[10], 64)
		lonp, _ := strconv.ParseFloat(line[11], 64)
		d, _ := strconv.ParseFloat(line[12], 64)
		dp, _ := strconv.ParseFloat(line[13], 64)

		AT = append(AT, ThrowRow{
			Date:     line[0],
			Course:   line[1],
			Round:    line[2],
			Hole:     line[3],
			Shot:     shot_num,
			Disc:     line[5],
			Lat1:     lat1,
			Lon1:     lon1,
			Lat2:     lat2,
			Lon2:     lon2,
			LatPin:   latp,
			LonPin:   lonp,
			Dist:     d,
			DistPin:  dp,
			ResThrow: line[14],
		})
	}
	return AT
}
