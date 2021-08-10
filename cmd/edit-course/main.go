package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

//go:embed templates/*
var templates embed.FS

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(templates, "templates/landing.html"))
	tmpl.Execute(w, nil)
}

func movepointHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	num_s := r.URL.Query()["num"]
	num, _ := strconv.Atoi(num_s[0])
	lat_s := r.URL.Query()["lat"]
	lat, _ := strconv.ParseFloat(lat_s[0], 64)
	lon_s := r.URL.Query()["lon"]
	lon, _ := strconv.ParseFloat(lon_s[0], 64)

	// dot := MovingDot{num, lat, lon}
	fmt.Println(num, lat, lon)

	time.Sleep(3)

	// os.Chdir("static")

	// ts := parseRoundRaw("round_raw.csv")
	// update_round_moving_point(ts, dot)
	// ts.WriteCSV()
	// cmnd := exec.Command("", "arg")
	//cmnd.Run() // and wait

	// os.Chdir("..")

	fmt.Fprint(w, "All ready bud.")
}

func main() {

	// in := os.Args[1]
	in := "kit_carson_small.json"

	crs := rnd.ParseCourseJSON(in)
	crs.DrawSummary()

	// make mapbox representation of course
	http.HandleFunc("/", homeHandler)

	// handle drags
	http.HandleFunc("/movepoint", movepointHandler)

	// handle renames

	// handle new pars

	// handle new clicks

	// serve up for local use

	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("."))))

	http.ListenAndServe(":8081", nil)

}
