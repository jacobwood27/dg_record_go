package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

type RoundRow struct {
	Date     string
	Course   string
	Round    string
	NumHoles int
	Par      int
	Total    int
	Score    int
}
type AllRounds []RoundRow

func (t RoundRow) asStrings() []string {
	return []string{
		t.Date,
		t.Course,
		t.Round,
		strconv.Itoa(t.NumHoles),
		strconv.Itoa(t.Par),
		strconv.Itoa(t.Total),
		strconv.Itoa(t.Score),
	}
}

func (a AllRounds) GetHeader() []string {
	return []string{
		"Date",
		"CourseID",
		"RoundID",
		"NumHoles",
		"Par",
		"Total",
		"Score",
	}
}

func GetRoundRow(rd rnd.Round) RoundRow {

	date := rd.ID[:10]
	courseID := rd.CourseID
	roundID := rd.ID

	course := rd.Course

	num_holes := 0
	total_shots := 0
	total_par := 0
	for _, r := range rd.Data {

		if r.Disc == "BASKET" {
			h := course.GetHole(r.HoleID)
			total_par = total_par + h.Par(r.TeeID, r.PinID)
			num_holes++
		} else {
			total_shots++
		}
	}

	return RoundRow{
		Date:     date,
		Course:   courseID,
		Round:    roundID,
		NumHoles: num_holes,
		Par:      total_par,
		Total:    total_shots,
		Score:    total_shots - total_par,
	}
}

func MakeAllRoundsCSV(rnds []rnd.Round) {

	var AR AllRounds
	for _, rd := range rnds {
		AR = append(AR, GetRoundRow(rd))
	}

	homedir, _ := os.UserHomeDir()
	at_csv := filepath.Join(homedir, ".discgolf", "stats", "all_rounds.csv")

	f, err := os.Create(at_csv)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write(AR.GetHeader())

	for _, t := range AR {
		w.Write(t.asStrings())
	}

}
