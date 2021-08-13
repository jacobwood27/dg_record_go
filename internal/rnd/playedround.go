package rnd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
)

type PlayedHole struct {
	HoleID string  `json:"hole"`
	TeeID  string  `json:"tee"`
	PinID  string  `json:"pin"`
	Par    int     `json:"par"`
	Stamps []Stamp `json:"stamps"`
}

type Round struct {
	ID      string       `json:"roundID"`
	CrsID   string       `json:"courseID"`
	CrsName string       `json:"courseName"`
	Holes   []PlayedHole `json:"holes"`
}

type RoundGEOJSON struct {
	Type     string            `json:"type"`
	Features []Feature         `json:"features"`
	Table    RoundScoreSummary `json:"table"`
}

func (ph PlayedHole) Score() int {
	return len(ph.Stamps) - 1
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

func processRoundRaw(ts Stamps, c Course) []PlayedHole {
	teeThresh := 10.0

	var phs []PlayedHole

	ph := PlayedHole{}
	h, t := inferHole(ts[0].Loc, c)
	ph.HoleID = h.ID
	ph.TeeID = t.ID
	ph.Stamps = append(ph.Stamps, Stamp{
		Loc:  t.Loc,
		Disc: ts[0].Disc,
	})
	for i, s := range ts[1:] {

		h_n, t_n := inferHole(s.Loc, c)
		d := dist(s.Loc, t_n.Loc)

		if d < teeThresh {
			// then add a pin to the previous hole
			p := inferPin(ts[i-1].Loc, h)
			ph.Stamps = append(ph.Stamps, Stamp{
				Loc:  p.Loc,
				Disc: "BASKET",
			})
			ph.PinID = p.ID
			ph.Par = h.Par(ph.TeeID, ph.PinID)
			phs = append(phs, ph)

			// and initialize a new one
			ph.Stamps = nil
			ph = PlayedHole{
				HoleID: h_n.ID,
				TeeID:  t_n.ID,
				Stamps: []Stamp{{
					Loc:  t_n.Loc,
					Disc: s.Disc,
				}},
			}

			h = h_n
			t = t_n
		} else {
			ph.Stamps = append(ph.Stamps, s)
		}

	}

	// and don't forget the last pin
	p := inferPin(ts[len(ts)-1].Loc, h)
	ph.Stamps = append(ph.Stamps, Stamp{
		Loc:  p.Loc,
		Disc: "BASKET",
	})
	ph.PinID = p.ID
	ph.Par = h.Par(ph.TeeID, ph.PinID)
	phs = append(phs, ph)

	return phs
}

func ParseRoundRaw(fileID string) (Round, Stamps) {
	ts := ReadRoundRawCSV(fileID)

	course := inferCourse(ts[0].Loc)

	rndID := fileID + "_-_" + course.ID

	PH := processRoundRaw(ts, course)

	r := Round{
		ID:      rndID,
		CrsID:   course.ID,
		CrsName: course.Name,
		Holes:   PH,
	}

	return r, ts
}

func (r Round) WriteRoundJSON() {
	file, _ := json.MarshalIndent(r, "", "	")
	_ = ioutil.WriteFile(r.ID+".json", file, 0644)
}

func ParseRoundJSON(filename string) Round {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var rnd Round
	json.Unmarshal([]byte(byteValue), &rnd)

	return rnd
}

func (r Round) MakeRoundVis() {

	var features []Feature
	for _, h := range r.Holes {
		holestring := [][]float64{}
		for _, s := range h.Stamps {
			geom := Geometry{
				Type:        "Point",
				Coordinates: []float64{s.Loc[1], s.Loc[0]},
			}
			f := Feature{
				Type: "Feature",
				Properties: Properties{
					Thing: "stamp",
				},
				Geometry: geom,
			}
			features = append(features, f)

			holestring = append(holestring, []float64{s.Loc[1], s.Loc[0]})
		}
		geom := Geometry{
			Type:        "LineString",
			Coordinates: holestring,
		}
		f := Feature{
			Type: "Feature",
			Properties: Properties{
				Thing: "holestring",
			},
			Geometry: geom,
		}
		features = append(features, f)
	}

	rgj := RoundGEOJSON{
		Type:     "FeatureCollection",
		Features: features,
	}

	file, _ := json.MarshalIndent(rgj, "", "	")
	_ = ioutil.WriteFile("round_vis.json", file, 0644)
}

type RoundTidyRow struct {
	RowNum int
	HoleID string
	TeeID  string
	PinID  string
	Lat    float64
	Lon    float64
	Disc   string
}

func (rtr RoundTidyRow) Loc() Loc {
	return Loc{rtr.Lat, rtr.Lon}
}

type RoundTidyTable []RoundTidyRow

func (rtt RoundTidyTable) getStamps() Stamps {
	s := Stamps{}
	for _, r := range rtt {
		s = append(s, Stamp{
			Loc:  r.Loc(),
			Disc: r.Disc,
		})
	}
	return s
}

type RoundTidy struct {
	ID         string
	CourseID   string
	CourseName string
	Course     Course
	Data       RoundTidyTable
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

func (rt *RoundTidy) DeleteLine(idx int) {
	ts := rt.Data.getStamps()
	ts = remove(ts, idx)
	c := rt.Course
	rtt := ProcessRoundRawWithPinStamp(ts, c)

	rt.Data = rtt
}

func (rt *RoundTidy) AddLine(idx int, lat float64, lon float64) {
	ts := rt.Data.getStamps()
	s := Stamp{
		Loc:  Loc{lat, lon},
		Disc: "",
	}
	ts = insert(ts, idx, s)
	c := rt.Course
	rtt := ProcessRoundRawWithPinStamp(ts, c)

	rt.Data = rtt
}

func (rt *RoundTidy) MovePoint(idx int, lat float64, lon float64) {
	ts := rt.Data.getStamps()
	ts[idx].Loc = Loc{lat, lon}
	c := rt.Course
	rtt := ProcessRoundRawWithPinStamp(ts, c)

	rt.Data = rtt
}

// func ProcessRoundRawNoPinStamp(ts Stamps, c Course) RoundTidyTable {
// 	teeThresh := 10.0

// 	var RTT RoundTidyTable

// 	h, t := inferHole(ts[0].Loc, c)
// 	rtr := RoundTidyRow{
// 		HoleID: h.ID,
// 		TeeID:  t.ID,
// 		PinID:  "",
// 		Lat:    t.Loc[0],
// 		Lon:    t.Loc[1],
// 		Disc:   ts[0].Disc,
// 	}
// 	RTT = append(RTT, rtr)

// 	for i, s := range ts[1:] {

// 		h_n, t_n := inferHole(s.Loc, c)
// 		d := dist(s.Loc, t_n.Loc)

// 		if d < teeThresh {
// 			// then add a pin to the previous hole
// 			p := inferPin(ts[i-1].Loc, h)

// 			rtr = RoundTidyRow{
// 				HoleID: h.ID,
// 				TeeID:  t.ID,
// 				PinID:  p.ID,
// 				Lat:    p.Loc[0],
// 				Lon:    p.Loc[1],
// 				Disc:   "BASKET",
// 			}
// 			RTT = append(RTT, rtr)

// 			// and add this drive
// 			h = h_n
// 			t = t_n
// 			rtr = RoundTidyRow{
// 				HoleID: h.ID,
// 				TeeID:  t.ID,
// 				PinID:  "",
// 				Lat:    t.Loc[0],
// 				Lon:    t.Loc[1],
// 				Disc:   s.Disc,
// 			}
// 			RTT = append(RTT, rtr)

// 		} else {
// 			rtr = RoundTidyRow{
// 				HoleID: h.ID,
// 				TeeID:  t.ID,
// 				PinID:  "",
// 				Lat:    s.Loc[0],
// 				Lon:    s.Loc[1],
// 				Disc:   s.Disc,
// 			}
// 			RTT = append(RTT, rtr)
// 		}

// 	}

// 	// and don't forget the last pin
// 	p := inferPin(ts[len(ts)-1].Loc, h)
// 	rtr = RoundTidyRow{
// 		HoleID: h.ID,
// 		TeeID:  t.ID,
// 		PinID:  p.ID,
// 		Lat:    p.Loc[0],
// 		Lon:    p.Loc[1],
// 		Disc:   "BASKET",
// 	}
// 	RTT = append(RTT, rtr)

// 	// and let's go through backwards and assign all the pins
// 	last_pin := p.ID
// 	for i := len(RTT) - 1; i >= 0; i-- {
// 		if RTT[i].PinID == "" {
// 			RTT[i].PinID = last_pin
// 		} else {
// 			last_pin = RTT[i].PinID
// 		}
// 		RTT[i].RowNum = i
// 	}

// 	return RTT
// }

func ProcessRoundRawWithPinStamp(ts Stamps, c Course) RoundTidyTable {
	teeThresh := 10.0
	pinThresh := 10.0
	driveThresh := 20.0

	var RTT RoundTidyTable

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
			RTT[len(RTT)-1].PinID = p.ID
		}

		rtr := RoundTidyRow{
			HoleID: h.ID,
			TeeID:  t.ID,
			PinID:  "",
			Lat:    s.Loc[0],
			Lon:    s.Loc[1],
			Disc:   s.Disc}

		if i == len(ts)-1 {
			rtr.PinID = p.ID
		}

		RTT = append(RTT, rtr)

	}

	// and let's go through backwards and assign all the pins
	last_pin := RTT[len(RTT)-1].PinID
	for i := len(RTT) - 1; i >= 0; i-- {
		if RTT[i].PinID == "" {
			RTT[i].PinID = last_pin
		} else {
			last_pin = RTT[i].PinID
		}
		RTT[i].RowNum = i
	}

	return RTT
}

func (rt RoundTidy) WriteCSV() {
	f, err := os.Create(rt.ID + "_tidy.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"RoundID: " + rt.ID})
	w.Write([]string{"CourseID: " + rt.CourseID})
	w.Write([]string{"CourseName: " + rt.CourseName})
	w.Write([]string{""})
	w.Write([]string{"row", "hole", "tee", "pin", "lat", "lon", "disc"})

	for _, r := range rt.Data {
		w.Write([]string{
			strconv.Itoa(r.RowNum),
			r.HoleID,
			r.TeeID,
			r.PinID,
			fmt.Sprintf("%f", r.Lat),
			fmt.Sprintf("%f", r.Lon),
			r.Disc,
		})
	}
}

func GetRoundTidy(fileID string) RoundTidy {
	ts := ReadRoundRawCSV(fileID)

	course := inferCourse(ts[0].Loc)

	rndID := fileID + "_-_" + course.ID

	RTT := ProcessRoundRawWithPinStamp(ts, course)

	RT := RoundTidy{
		ID:         rndID,
		CourseID:   course.ID,
		CourseName: course.Name,
		Data:       RTT,
		Course:     course,
	}

	return RT
}

// func remove(slice RoundTidyTable, s int) RoundTidyTable {
// 	return append(slice[:s], slice[s+1:]...)
// }

// func (RT RoundTidy) ReprocessRoundTidy() {
// 	teeThresh := 3.0

// 	old := RT.Data
// 	c := RT.Course

// 	var RTT RoundTidyTable

// 	h, t := inferHole(old[0].Loc(), c)
// 	rtr := RoundTidyRow{
// 		HoleID: h.ID,
// 		TeeID:  t.ID,
// 		PinID:  "",
// 		Lat:    t.Loc[0],
// 		Lon:    t.Loc[1],
// 		Disc:   old[0].Disc,
// 	}
// 	RTT = append(RTT, rtr)

// 	for i, s := range ts[1:] {

// 		h_n, t_n := inferHole(s.Loc, c)
// 		d := dist(s.Loc, t_n.Loc)

// 		if d < teeThresh {
// 			// then add a pin to the previous hole
// 			p := inferPin(ts[i-1].Loc, h)

// 			rtr = RoundTidyRow{
// 				HoleID: h.ID,
// 				TeeID:  t.ID,
// 				PinID:  p.ID,
// 				Lat:    p.Loc[0],
// 				Lon:    p.Loc[1],
// 				Disc:   "BASKET",
// 			}
// 			RTT = append(RTT, rtr)

// 			// and add this drive
// 			h = h_n
// 			t = t_n
// 			rtr = RoundTidyRow{
// 				HoleID: h.ID,
// 				TeeID:  t.ID,
// 				PinID:  "",
// 				Lat:    t.Loc[0],
// 				Lon:    t.Loc[1],
// 				Disc:   s.Disc,
// 			}
// 			RTT = append(RTT, rtr)

// 		} else {
// 			rtr = RoundTidyRow{
// 				HoleID: h.ID,
// 				TeeID:  t.ID,
// 				PinID:  "",
// 				Lat:    s.Loc[0],
// 				Lon:    s.Loc[1],
// 				Disc:   s.Disc,
// 			}
// 			RTT = append(RTT, rtr)
// 		}

// 	}

// 	// and don't forget the last pin
// 	p := inferPin(ts[len(ts)-1].Loc, h)
// 	rtr = RoundTidyRow{
// 		HoleID: h.ID,
// 		TeeID:  t.ID,
// 		PinID:  p.ID,
// 		Lat:    p.Loc[0],
// 		Lon:    p.Loc[1],
// 		Disc:   "BASKET",
// 	}
// 	RTT = append(RTT, rtr)

// 	// and let's go through backwards and assign all the pins
// 	last_pin := p.ID
// 	for i := len(RTT) - 1; i >= 0; i-- {
// 		if RTT[i].PinID == "" {
// 			RTT[i].PinID = last_pin
// 		} else {
// 			last_pin = RTT[i].PinID
// 		}
// 		RTT[i].RowNum = i
// 	}

// 	return RTT
// }

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

func (rt RoundTidy) DrawSummary() {

	var features []Feature
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
				Thing: "stamp",
				Name:  strconv.Itoa(r.RowNum),
			},
			Geometry: geom,
		}
		features = append(features, f)
	}

	cur_hole = rt.Data[0].HoleID
	for i, r := range rt.Data {
		if i == 0 {
			continue
		}

		if r.HoleID != cur_hole {
			cur_hole = r.HoleID
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
					Thing: "throw",
					Name:  strconv.Itoa(i),
				},
				Geometry: geom,
			}
			features = append(features, f)
		}

	}

	rgj := RoundGEOJSON{
		Type:     "FeatureCollection",
		Features: features,
		Table:    RSS,
	}

	file, _ := json.MarshalIndent(rgj, "", "	")
	_ = ioutil.WriteFile("round_vis.json", file, 0644)
}
