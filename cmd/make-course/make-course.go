package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

type csvstr struct {
	hole      string
	variation string
	lat       float64
	lon       float64
}
type csvFile []csvstr

func readCSV(fname string) csvFile {
	var l []csvstr

	f, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	_, err = r.Read() // Burn the header line
	if err != nil {
		log.Fatal(err)
	}

	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		lat, _ := strconv.ParseFloat(line[2], 64)
		lon, _ := strconv.ParseFloat(line[3], 64)
		l = append(l, csvstr{
			hole:      line[0],
			variation: line[1],
			lat:       lat,
			lon:       lon,
		})
	}

	return csvFile(l)
}

func makeCourseJSON(id string, name string, tees csvFile, pins csvFile) rnd.Course {

	crsLoc := rnd.Loc{tees[0].lat, tees[0].lon}

	cur_hole := ""
	var ts []rnd.Tee
	var ps []rnd.Pin
	var pars []rnd.Par
	var holes []rnd.Hole
	for i, t := range tees {
		if t.hole != cur_hole {

			if i > 0 {
				holes = append(holes, rnd.Hole{
					ID:   cur_hole,
					Tees: ts,
					Pins: ps,
					Pars: pars,
				})
			}

			cur_hole = t.hole
			ts = nil
			ps = nil
			pars = nil
		}

		ts = append(ts, rnd.Tee{
			ID:  t.variation,
			Loc: rnd.Loc{t.lat, t.lon},
		})

		for _, p := range pins {
			if p.hole == t.hole {
				ps = append(ps, rnd.Pin{
					ID:  p.variation,
					Loc: rnd.Loc{p.lat, p.lon},
				})

				pars = append(pars, rnd.Par{
					Tee: t.variation,
					Pin: p.variation,
					Par: 3,
				})

			}
		}
	}
	holes = append(holes, rnd.Hole{
		ID:   cur_hole,
		Tees: ts,
		Pins: ps,
		Pars: pars,
	})

	return rnd.Course{
		ID:    id,
		Name:  name,
		Loc:   crsLoc,
		Holes: holes,
	}
}

func main() {

	tees := readCSV("tees.csv")
	pins := readCSV("pins.csv")

	courseID := os.Args[1]
	courseName := os.Args[2]

	cJSON := makeCourseJSON(courseID, courseName, tees, pins)

	file, _ := json.MarshalIndent(cJSON, "", "	")

	_ = ioutil.WriteFile(courseID+".json", file, 0644)
}
