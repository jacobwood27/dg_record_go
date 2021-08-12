package rnd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
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

type Properties struct {
	Thing    string `json:"thing"`
	DiscName string `json:"disc_name"`
	Name     string `json:"name"`
	Par      int    `json:"par"`
}

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

type Feature struct {
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
}

type SummaryTableRow struct {
	Hole string  `json:"hole"`
	Tee  string  `json:"tee"`
	Pin  string  `json:"pin"`
	Dist float64 `json:"dist"`
	Par  int     `json:"par"`
}

type SummaryTable []SummaryTableRow

type CourseGEOJSON struct {
	Type     string       `json:"type"`
	Features []Feature    `json:"features"`
	Table    SummaryTable `json:"table"`
}

func (c Course) DrawSummary() {

	var features []Feature
	for _, h := range c.Holes {
		for _, t := range h.Tees {
			geom := Geometry{
				Type:        "Point",
				Coordinates: []float64{t.Loc[1], t.Loc[0]},
			}
			f := Feature{
				Type: "Feature",
				Properties: Properties{
					Thing: "tee",
					Name:  h.ID + "_" + t.ID,
				},
				Geometry: geom,
			}
			features = append(features, f)

			for _, p := range h.Pins {
				geom := Geometry{
					Type:        "LineString",
					Coordinates: [][]float64{{t.Loc[1], t.Loc[0]}, {p.Loc[1], p.Loc[0]}},
				}
				f := Feature{
					Type: "Feature",
					Properties: Properties{
						Thing: "tee->pin",
						Par:   h.Par(t.ID, p.ID),
						Name:  h.ID + "_" + t.ID + "->" + p.ID,
					},
					Geometry: geom,
				}
				features = append(features, f)
			}
		}

		for _, p := range h.Pins {
			geom := Geometry{
				Type:        "Point",
				Coordinates: []float64{p.Loc[1], p.Loc[0]},
			}
			f := Feature{
				Type: "Feature",
				Properties: Properties{
					Thing: "pin",
					Name:  h.ID + "_" + p.ID,
				},
				Geometry: geom,
			}
			features = append(features, f)
		}
	}

	var tRows []SummaryTableRow
	for _, h := range c.Holes {
		for i, t := range h.Tees {
			for j, p := range h.Pins {
				holename := ""
				teename := ""
				pinname := p.ID
				if i == 0 && j == 0 {
					holename = h.ID
				}
				if j == 0 {
					teename = t.ID
				}
				par := h.Par(t.ID, p.ID)

				tRows = append(tRows, SummaryTableRow{
					Hole: holename,
					Tee:  teename,
					Pin:  pinname,
					Dist: math.Round(dist(t.Loc, p.Loc) * 3.28),
					Par:  par,
				})
			}
		}
	}

	cgj := CourseGEOJSON{
		Type:     "FeatureCollection",
		Features: features,
		Table:    SummaryTable(tRows),
	}

	file, _ := json.MarshalIndent(cgj, "", "	")
	_ = ioutil.WriteFile("course_vis.json", file, 0644)
}

type Throw struct {
	Loc1 Loc
	Loc2 Loc
	Disc string
}

type PlayedHole struct {
	HoleID string
	TeeID  string
	PinID  string
	Throws []Throw
	Score  int
	Par    int
}

type Round struct {
	ID      string
	CrsID   string
	CrsName string
	Holes   []PlayedHole
}

func (h Hole) Par(tID string, pID string) int {
	for _, p := range h.Pars {
		if p.Tee == tID && p.Pin == pID {
			return p.Par
		}
	}
	return 99
}

func ParseCourseJSON(filename string) Course {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var crs Course
	json.Unmarshal([]byte(byteValue), &crs)

	return crs
}

func (c Course) SaveCourseJSON() {
	file, _ := json.MarshalIndent(c, "", "	")
	_ = ioutil.WriteFile(c.ID+".json", file, 0644)
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func inferCourse(l Loc) Course {
	crs_path := path.Join(RootDir(), "..", "data", "courses")
	files, e := os.ReadDir(crs_path)
	if e != nil {
		panic(e)
	}

	best_c := Course{}
	best_dist := 999.9
	for _, file := range files {
		c := ParseCourseJSON(path.Join(crs_path, file.Name()))
		this_dist := dist(c.Loc, l)

		if this_dist < best_dist {
			best_c = c
			best_dist = this_dist
		}

	}

	return best_c
}

func inferHole(l Loc, c Course) (Hole, Tee) {
	bestH := Hole{}
	bestT := Tee{}
	best_dist := 999.9

	for _, h := range c.Holes {
		fmt.Println("here")
		for _, t := range h.Tees {
			this_dist := dist(t.Loc, l)
			fmt.Println(this_dist)

			if this_dist < best_dist {
				bestH = h
				bestT = t
				best_dist = this_dist
			}
		}
	}

	return bestH, bestT
}

func inferPlayedHole(ts Throwstamps, tee Tee, h Hole, c Course) (PlayedHole, Throwstamps, Hole, Tee) {
	pinThresh := 10.0
	teeThresh := 10.0
	driveThresh := 10.0

	var throws []Throw

	for i, t := range ts {

		if i > 1 {
			throws = append(throws, Throw{
				Loc1: ts[i-1].Loc(),
				Loc2: ts[i].Loc(),
				Disc: ts[i-1].Disc,
			})
		}

		bestP := Pin{}
		best_dist := 99.9
		for _, p := range h.Pins {
			d_p := dist(t.Loc(), p.Loc) // distance from this stamp to this pin

			if d_p < best_dist {
				best_dist = d_p
				bestP = p
			}
		}

		if best_dist < pinThresh {
			hole, tee := inferHole(ts[i+1].Loc(), c)

			nextT_dist := dist(tee.Loc, ts[i+1].Loc())
			if nextT_dist < teeThresh {
				nextD_dist := dist(ts[i+1].Loc(), ts[i+2].Loc())

				if nextD_dist > driveThresh {
					//winner winner!
					ph := PlayedHole{
						HoleID: h.ID,
						TeeID:  tee.ID,
						PinID:  bestP.ID,
						Throws: throws,
						Score:  len(throws),
						Par:    h.Par(tee.ID, bestP.ID),
					}

					ts = ts[i:]

					return ph, ts, hole, tee
				}
			}
		}
	}
	return PlayedHole{}, ts, Hole{}, Tee{}
}

func ParseRoundRaw(filename string) Round {
	ts := ReadRoundRawCSV(filename)

	course := inferCourse(ts[0].Loc())
	var PH []PlayedHole

	hole, tee := inferHole(ts[0].Loc(), course)
	ph, ts, nextH, nextT := inferPlayedHole(ts, tee, hole, course)
	PH = append(PH, ph)

	fmt.Println(hole, tee, nextH, nextT)

	rndID := ts[0].Time.Format("2006-01-02-15-04-05") + "_-_" + course.ID

	return Round{
		ID:      rndID,
		CrsID:   course.ID,
		CrsName: course.Name,
		Holes:   PH,
	}
}
