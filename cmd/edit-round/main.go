package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

//go:embed templates/*
var templates embed.FS

// var round rnd.Round
// var ts rnd.Stamps
var rt rnd.RoundTidy

// var crs rnd.Course

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(templates, "templates/edit_round.html"))
	tmpl.Execute(w, nil)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	name_s := r.URL.Query()["name"]
	name, _ := strconv.Atoi(name_s[0])

	rt.DeleteLine(name)
	rt.DrawSummary()

}

func addHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	name_s := r.URL.Query()["name"]
	name, _ := strconv.Atoi(name_s[0])
	lat_s := r.URL.Query()["lat"]
	lat, _ := strconv.ParseFloat(lat_s[0], 64)
	lon_s := r.URL.Query()["lon"]
	lon, _ := strconv.ParseFloat(lon_s[0], 64)

	rt.AddLine(name, lat, lon)
	rt.DrawSummary()

}

func movepointHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	name_s := r.URL.Query()["name"]
	name, _ := strconv.Atoi(name_s[0])
	lat_s := r.URL.Query()["lat"]
	lat, _ := strconv.ParseFloat(lat_s[0], 64)
	lon_s := r.URL.Query()["lon"]
	lon, _ := strconv.ParseFloat(lon_s[0], 64)

	rt.MovePoint(name, lat, lon)
	rt.DrawSummary()

}

// func movepointHandler(w http.ResponseWriter, r *http.Request) {

// 	fmt.Println(r.URL)

// 	name_s := r.URL.Query()["name"]
// 	name := name_s[0]
// 	lat_s := r.URL.Query()["lat"]
// 	lat, _ := strconv.ParseFloat(lat_s[0], 64)
// 	lon_s := r.URL.Query()["lon"]
// 	lon, _ := strconv.ParseFloat(lon_s[0], 64)

// 	for _, h := range crs.Holes {
// 		for _, t := range h.Tees {
// 			if h.ID+"_"+t.ID == name {
// 				t.Loc[0] = lat
// 				t.Loc[1] = lon
// 				crs.DrawSummary()
// 				return
// 			}
// 		}
// 	}

// 	for _, h := range crs.Holes {
// 		for _, p := range h.Pins {
// 			if h.ID+"_"+p.ID == name {
// 				p.Loc[0] = lat
// 				p.Loc[1] = lon

// 				crs.DrawSummary()
// 				return
// 			}
// 		}
// 	}

// 	fmt.Println("Shouldn't be here")

// }

func main() {
	// in := os.Args[1]
	matches, _ := filepath.Glob("./*_raw.csv")
	m := matches[0]
	fID := m[:len(m)-8]

	rt = rnd.GetRoundTidy(fID)
	rt.DrawSummary()

	// crs = rnd.ParseCourseJSON(in)
	// crs.DrawSummary()

	// round.MakeRoundVis()

	// // make mapbox representation of course
	http.HandleFunc("/", homeHandler)

	// // handle moving points
	http.HandleFunc("/movepoint", movepointHandler)

	// handle deleting points
	http.HandleFunc("/deletepoint", deleteHandler)

	// // handle adding points by splitting a line
	http.HandleFunc("/addpoint", addHandler)

	// // serve local files as if they were in /data
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("."))))

	// // serve up for local use
	http.ListenAndServe(":8081", nil)
}
