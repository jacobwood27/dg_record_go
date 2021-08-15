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

type HoleRow struct {
	Date      string
	Course    string
	Round     string
	Hole      string
	Par       int
	Tee       string
	Pin       string
	Length    float64
	DriveDist float64
	MakeDist  float64
	Score     int
	Res       int
	ResName   string
}
type AllHoles []HoleRow

func (t HoleRow) asStrings() []string {
	return []string{
		t.Date,
		t.Course,
		t.Round,
		t.Hole,
		strconv.Itoa(t.Par),
		t.Tee,
		t.Pin,
		fmt.Sprintf("%f", t.Length),
		fmt.Sprintf("%f", t.DriveDist),
		fmt.Sprintf("%f", t.MakeDist),
		strconv.Itoa(t.Score),
		strconv.Itoa(t.Res),
		t.ResName,
	}
}

func (a AllHoles) GetHeader() []string {
	return []string{
		"Date",
		"CourseID",
		"RoundID",
		"HoleID",
		"Par",
		"Tee",
		"Pin",
		"Length",
		"DriveDist",
		"MakeDist",
		"Score",
		"Result",
		"ResultLabel",
	}
}

var scoreLabels = map[int]string{
	-3: "ALBATROSS",
	-2: "EAGLE",
	-1: "BIRDIE",
	0:  "PAR",
	1:  "BOGEY",
	2:  "DOUBLE BOGEY",
	3:  "TRIPLE BOGEY",
	4:  "QUADRUPLE BOGEY",
	5:  "BAD NEWS",
}

func GetRoundHoles(rd rnd.Round) AllHoles {

	var AH AllHoles

	date := rd.ID[:10]
	courseID := rd.CourseID
	roundID := rd.ID
	course := rd.Course

	p := rnd.Pin{}
	t := rnd.Tee{}
	h := rnd.Hole{}

	cur_hole := ""
	shot_num := 1
	for i, r := range rd.Data {

		if r.HoleID != cur_hole {
			h = course.GetHole(r.HoleID)
			p = h.GetPin(r.PinID)
			t = h.GetTee(r.TeeID)

			nr := rd.Data[i+1]

			AH = append(AH, HoleRow{
				Date:      date,
				Course:    courseID,
				Round:     roundID,
				Hole:      h.ID,
				Par:       h.Par(t.ID, p.ID),
				Tee:       t.ID,
				Pin:       p.ID,
				Length:    math.Round(3.281 * rnd.Dist(t.Loc, p.Loc)),
				DriveDist: math.Round(3.281 * rnd.Dist(t.Loc, nr.Loc())),
				MakeDist:  0.0,
				Score:     0,
				Res:       0,
				ResName:   ""})

			cur_hole = r.HoleID
			shot_num = 1
		} else {
			shot_num++
		}

		if r.Disc == "BASKET" {
			// then we finish writing this guy
			pr := rd.Data[i-1]
			AH[len(AH)-1].MakeDist = math.Round(3.281 * rnd.Dist(pr.Loc(), p.Loc))
			AH[len(AH)-1].Score = shot_num - 1
			AH[len(AH)-1].Res = AH[len(AH)-1].Score - AH[len(AH)-1].Par
			AH[len(AH)-1].ResName = scoreLabels[AH[len(AH)-1].Res]
		}
	}

	return AH
}

func MakeAllHolesCSV(rnds []rnd.Round) {

	var AH AllHoles
	for _, rd := range rnds {
		AH = append(AH, GetRoundHoles(rd)...)
	}

	homedir, _ := os.UserHomeDir()
	at_csv := filepath.Join(homedir, ".discgolf", "stats", "all_holes.csv")

	f, err := os.Create(at_csv)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write(AH.GetHeader())

	for _, t := range AH {
		w.Write(t.asStrings())
	}

}
