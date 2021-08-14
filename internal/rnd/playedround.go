package rnd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
)

type RoundRow struct {
	RowNum int     `json:"-"`
	HoleID string  `json:"hole"`
	TeeID  string  `json:"tee"`
	PinID  string  `json:"pin"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Disc   string  `json:"disc"`
}

type RoundTable []RoundRow

type Round struct {
	ID         string     `json:"roundID"`
	CourseID   string     `json:"courseID"`
	CourseName string     `json:"courseName"`
	Course     Course     `json:"-"`
	Data       RoundTable `json:"roundData"`
	Notes      string     `json:"notes"`
}

type HoleScoreSummary struct {
	Hole   string  `json:"hole"`
	Tee    string  `json:"tee"`
	Pin    string  `json:"pin"`
	Dist   float64 `json:"dist"`
	Par    int     `json:"par"`
	Score  int     `json:"score"`
	Result int     `json:"res"`
	Total  int     `json:"tot"`
}
type RoundScoreSummary []HoleScoreSummary

type Disc struct {
	ID    string `json:"id"`
	Image string `json:"image"`
}

type RoundGEOJSON struct {
	Type     string            `json:"type"`
	Discs    []Disc            `json:"discs"`
	Features []Feature         `json:"features"`
	Table    RoundScoreSummary `json:"table"`
}

func (rr RoundRow) Loc() Loc {
	return Loc{rr.Lat, rr.Lon}
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func inferHole(l Loc, c Course) (Hole, Tee) {
	bestH := Hole{}
	bestT := Tee{}
	best_dist := 999.9

	for _, h := range c.Holes {
		for _, t := range h.Tees {

			this_dist := dist(t.Loc, l)

			if this_dist < best_dist {
				bestH = h
				bestT = t
				best_dist = this_dist
			}
		}
	}

	return bestH, bestT
}

func inferPin(l Loc, h Hole) Pin {
	bestP := Pin{}
	best_dist := 9999.9

	for _, p := range h.Pins {
		this_dist := dist(p.Loc, l)

		if this_dist < best_dist {
			bestP = p
			best_dist = this_dist
		}
	}

	return bestP
}

func (rt RoundTable) getStamps() Stamps {
	s := Stamps{}
	for _, r := range rt {
		s = append(s, Stamp{
			Loc:  r.Loc(),
			Disc: r.Disc,
		})
	}
	return s
}

func remove(s Stamps, i int) Stamps {
	s = append(s[:i], s[i+1:]...)
	return s
}
func insert(s Stamps, i int, value Stamp) Stamps {
	s = append(s[:i+1], s[i:]...)
	s[i] = value
	return s
}

func (r *Round) DeleteLine(idx int) {
	ts := r.Data.getStamps()
	ts = remove(ts, idx)
	c := r.Course
	rt := ProcessStamps(ts, c)

	r.Data = rt
}

func (r *Round) AddLine(idx int, lat float64, lon float64) {
	ts := r.Data.getStamps()
	s := Stamp{
		Loc:  Loc{lat, lon},
		Disc: "UNDEFINED",
	}
	ts = insert(ts, idx, s)
	c := r.Course
	rt := ProcessStamps(ts, c)

	r.Data = rt
}

func (r *Round) MovePoint(idx int, lat float64, lon float64) {
	ts := r.Data.getStamps()
	ts[idx].Loc = Loc{lat, lon}
	c := r.Course
	rt := ProcessStamps(ts, c)

	r.Data = rt
}

func ProcessStamps(ts Stamps, c Course) RoundTable {
	teeThresh := 10.0
	pinThresh := 10.0
	driveThresh := 20.0

	var RT RoundTable

	h, t := inferHole(ts[0].Loc, c)

	for i, s := range ts {

		h_n, t_n := inferHole(s.Loc, c)
		d_tee := dist(s.Loc, t_n.Loc)

		d_pin := 999.9
		p := Pin{}
		if i > 0 {
			p = inferPin(ts[i-1].Loc, h)
			d_pin = dist(ts[i-1].Loc, p.Loc)
		}

		nextDriveDist := 0.0
		if i < len(ts)-1 {
			nextDriveDist = dist(ts[i+1].Loc, ts[i].Loc)
		}

		if i > 0 && d_tee < teeThresh && d_pin < pinThresh && nextDriveDist > driveThresh {
			// then roll to the next hole
			h = h_n
			t = t_n

			// and assign previous pin
			RT[len(RT)-1].PinID = p.ID
		}

		rr := RoundRow{
			HoleID: h.ID,
			TeeID:  t.ID,
			PinID:  "",
			Lat:    s.Loc[0],
			Lon:    s.Loc[1],
			Disc:   s.Disc}

		if i == len(ts)-1 {
			rr.PinID = p.ID
		}

		RT = append(RT, rr)

	}

	// and let's go through backwards and assign all the pins
	last_pin := RT[len(RT)-1].PinID
	for i := len(RT) - 1; i >= 0; i-- {
		if RT[i].PinID == "" {
			RT[i].PinID = last_pin
		} else {
			last_pin = RT[i].PinID
		}
		RT[i].RowNum = i
	}

	return RT
}

func (r Round) WriteCSV() {
	f, err := os.Create(r.ID + ".csv")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"RoundID: " + r.ID})
	w.Write([]string{"CourseID: " + r.CourseID})
	w.Write([]string{"CourseName: " + r.CourseName})
	w.Write([]string{"Notes: " + r.Notes})
	w.Write([]string{""})
	w.Write([]string{"row", "hole", "tee", "pin", "lat", "lon", "disc"})

	for _, l := range r.Data {
		w.Write([]string{
			strconv.Itoa(l.RowNum),
			l.HoleID,
			l.TeeID,
			l.PinID,
			fmt.Sprintf("%f", l.Lat),
			fmt.Sprintf("%f", l.Lon),
			l.Disc,
		})
	}
}

func (r Round) WriteJSON() {
	file, _ := json.MarshalIndent(r, "", "	")
	_ = ioutil.WriteFile(r.ID+".json", file, 0644)
}

func (r Round) Cleanup() {

	cur_hole := ""
	for i, l := range r.Data {
		if l.HoleID != cur_hole {
			h := r.Course.GetHole(l.HoleID)
			t := h.GetTee(l.TeeID)
			r.Data[i].Lat = t.Loc[0]
			r.Data[i].Lon = t.Loc[1]
			cur_hole = l.HoleID
		}

		if i == len(r.Data)-1 || l.HoleID != r.Data[i+1].HoleID {
			r.Data[i].Disc = "BASKET"
			h := r.Course.GetHole(l.HoleID)
			p := h.GetPin(l.PinID)
			r.Data[i].Lat = p.Loc[0]
			r.Data[i].Lon = p.Loc[1]
		}
	}

}

func GetRound(ts Stamps, fileID string) Round {

	course := inferCourse(ts[0].Loc)

	rndID := fileID + "_-_" + course.ID

	fmt.Println(rndID)
	RT := ProcessStamps(ts, course)

	R := Round{
		ID:         rndID,
		CourseID:   course.ID,
		CourseName: course.Name,
		Data:       RT,
		Course:     course,
	}

	return R
}

type discCSVRow struct {
	Name    string
	ID      string
	Plastic string
	Mass    float64
	Image   string
}

type discCSV []discCSVRow

func GetDiscs() discCSV {
	homedir, _ := os.UserHomeDir()
	filename := filepath.Join(homedir, ".discgolf", "discs", "discs.csv")

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Read() //burn header

	var ts discCSV
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		mass, _ := strconv.ParseFloat(line[3], 64)
		ts = append(ts, discCSVRow{line[0], line[1], line[2], mass, line[4]})
	}
	return ts
}

func (rt Round) DrawSummary() {

	var features []Feature

	all_discs := GetDiscs()
	used_discs := []Disc{{"UNDEFINED", "blank.png"}, {"BASKET", "basket.png"}}

	for _, d := range all_discs {
		used_discs = append(used_discs, Disc{
			ID:    d.Name,
			Image: d.Image,
		})
	}

	c := rt.Course

	var RSS RoundScoreSummary
	cur_hole := rt.Data[0].HoleID
	cur_tot := 0
	hole_tot := -2
	for i, r := range rt.Data {
		hole_tot++

		if i > 0 && r.HoleID != cur_hole {
			rp := rt.Data[i-1]
			h := c.GetHole(rp.HoleID)
			t := h.GetTee(rp.TeeID)
			p := h.GetPin(rp.PinID)
			result := hole_tot - h.Par(t.ID, p.ID)
			cur_tot = cur_tot + result
			RSS = append(RSS, HoleScoreSummary{
				Hole:   h.ID,
				Tee:    t.ID,
				Pin:    p.ID,
				Dist:   math.Round(3.281 * dist(t.Loc, p.Loc)),
				Par:    h.Par(t.ID, p.ID),
				Score:  hole_tot,
				Result: result,
				Total:  cur_tot,
			})

			cur_hole = r.HoleID
			hole_tot = -1
		}
	}
	hole_tot++
	rp := rt.Data[len(rt.Data)-1]
	h := c.GetHole(rp.HoleID)
	t := h.GetTee(rp.TeeID)
	p := h.GetPin(rp.PinID)
	result := hole_tot - h.Par(t.ID, p.ID)
	cur_tot = cur_tot + result
	RSS = append(RSS, HoleScoreSummary{
		Hole:   h.ID,
		Tee:    t.ID,
		Pin:    p.ID,
		Dist:   math.Round(3.281 * dist(t.Loc, p.Loc)),
		Par:    h.Par(t.ID, p.ID),
		Score:  hole_tot,
		Result: result,
		Total:  cur_tot,
	})

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

	for _, r := range rt.Data {
		geom := Geometry{
			Type:        "Point",
			Coordinates: []float64{r.Lon, r.Lat},
		}
		f := Feature{
			Type: "Feature",
			Properties: Properties{
				Thing:    "stamp",
				Name:     strconv.Itoa(r.RowNum),
				DiscName: r.Disc,
			},
			Geometry: geom,
		}
		features = append(features, f)
	}

	cur_hole = rt.Data[0].HoleID
	cur_hole_i := 0
	for i, r := range rt.Data {
		if i == 0 {
			continue
		}

		if r.HoleID != cur_hole {
			cur_hole = r.HoleID
			cur_hole_i++
			geom := Geometry{
				Type:        "LineString",
				Coordinates: [][]float64{{rt.Data[i-1].Lon, rt.Data[i-1].Lat}, {rt.Data[i].Lon, rt.Data[i].Lat}},
			}
			f := Feature{
				Type: "Feature",
				Properties: Properties{
					Thing: "walk",
					Name:  strconv.Itoa(i),
				},
				Geometry: geom,
			}
			features = append(features, f)
		} else {
			geom := Geometry{
				Type:        "LineString",
				Coordinates: [][]float64{{rt.Data[i-1].Lon, rt.Data[i-1].Lat}, {rt.Data[i].Lon, rt.Data[i].Lat}},
			}
			f := Feature{
				Type: "Feature",
				Properties: Properties{
					Thing:  "throw",
					Name:   strconv.Itoa(i),
					Result: RSS[cur_hole_i].Result,
				},
				Geometry: geom,
			}
			features = append(features, f)
		}

	}

	rgj := RoundGEOJSON{
		Type:     "FeatureCollection",
		Discs:    used_discs,
		Features: features,
		Table:    RSS,
	}

	file, _ := json.MarshalIndent(rgj, "", "	")
	_ = ioutil.WriteFile("round_vis.json", file, 0644)
}
