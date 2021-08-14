package rnd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
)

type Properties struct {
	Thing    string `json:"thing"`
	DiscName string `json:"disc_name"`
	Name     string `json:"name"`
	Par      int    `json:"par"`
	Result   int    `json:"res"`
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
