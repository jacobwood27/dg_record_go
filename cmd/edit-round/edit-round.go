package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

//go:embed templates/*
var templates embed.FS

//  //go:embed ../../data/courses
// var courses embed.FS

var rt rnd.Round

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

func saveHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	rt.Cleanup()
	rt.DrawSummary()

	rt.WriteCSV()

	fmt.Println("Round saved.")
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)

}

func snapHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	rt.Cleanup()
	rt.DrawSummary()

	http.Redirect(w, r, "/", http.StatusPermanentRedirect)

}

func main() {
	// in := os.Args[1]
	// matches, _ := filepath.Glob("./*_raw.csv")
	// m := matches[0]
	// fID := m[:len(m)-8]

	// ts_csv := "/home/woojac/.discgolf/rounds/2021-08-03-13-25_sunset_park/ts.csv"
	// rec_gpx := "/home/woojac/.discgolf/rounds/2021-08-03-13-25_sunset_park/rec.gpx"

	// ts, fID := rnd.GetRoundRaw(ts_csv, rec_gpx)

	// rt = rnd.GetRound(ts, fID)
	cwd, _ := os.Getwd()
	path := filepath.Base(cwd)
	fmt.Println(path)
	rt = rnd.ReadRoundCSV(path + ".csv")

	rt.DrawSummary()

	// // make mapbox representation of course
	http.HandleFunc("/", homeHandler)

	// Save as .csv
	http.HandleFunc("/save", saveHandler)

	// Snap to tees+pins
	http.HandleFunc("/snap", snapHandler)

	// // handle moving points
	http.HandleFunc("/movepoint", movepointHandler)

	// handle deleting points
	http.HandleFunc("/deletepoint", deleteHandler)

	// // handle adding points by splitting a line
	http.HandleFunc("/addpoint", addHandler)

	// // serve local files as if they were in /data
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("."))))

	// // serve local files as if they were in /coursedata
	// http.Handle("/coursedata/", http.StripPrefix("/coursedata/", http.FileServer(http.FS(courses))))

	// // serve up for local use
	fmt.Println("Edit the round at 0.0.0.0:8082")
	http.ListenAndServe(":8082", nil)
}
