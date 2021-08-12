package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

//go:embed templates/*
var templates embed.FS

var crs rnd.Course

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(templates, "templates/landing.html"))
	tmpl.Execute(w, nil)
}

func movepointHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	name_s := r.URL.Query()["name"]
	name := name_s[0]
	lat_s := r.URL.Query()["lat"]
	lat, _ := strconv.ParseFloat(lat_s[0], 64)
	lon_s := r.URL.Query()["lon"]
	lon, _ := strconv.ParseFloat(lon_s[0], 64)

	for _, h := range crs.Holes {
		for _, t := range h.Tees {
			if h.ID+"_"+t.ID == name {
				t.Loc[0] = lat
				t.Loc[1] = lon
				crs.DrawSummary()
				return
			}
		}
	}

	for _, h := range crs.Holes {
		for _, p := range h.Pins {
			if h.ID+"_"+p.ID == name {
				p.Loc[0] = lat
				p.Loc[1] = lon

				crs.DrawSummary()
				return
			}
		}
	}

	fmt.Println("Shouldn't be here")

}

func textHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	orig_s := r.URL.Query()["orig"]
	orig := orig_s[0]
	new_s := r.URL.Query()["new"]
	newN := new_s[0]

	for i, h := range crs.Holes {
		for j, t := range h.Tees {
			if h.ID+"_"+t.ID == orig {
				s := strings.Split(newN, "_")
				crs.Holes[i].Tees[j].ID = s[1]
				crs.DrawSummary()
				return
			}
		}
	}

	for i, h := range crs.Holes {
		for j, p := range h.Pins {
			if h.ID+"_"+p.ID == orig {
				s := strings.Split(newN, "_")
				crs.Holes[i].Pins[j].ID = s[1]
				crs.DrawSummary()
				return
			}
		}
	}

	fmt.Println("Shouldn't be here")

}

func newPOIHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	hole_s := r.URL.Query()["hole"]
	hole := hole_s[0]
	tee_s := r.URL.Query()["tee"]
	tee := tee_s[0]
	pin_s := r.URL.Query()["pin"]
	pin := pin_s[0]
	lat_s := r.URL.Query()["lat"]
	lat, _ := strconv.ParseFloat(lat_s[0], 64)
	lon_s := r.URL.Query()["lon"]
	lon, _ := strconv.ParseFloat(lon_s[0], 64)

	// Add new hole if needed
	new_hole := true
	for _, h := range crs.Holes {
		if h.ID == hole {
			new_hole = false
			break
		}
	}
	if new_hole {
		crs.Holes = append(crs.Holes, rnd.Hole{
			ID: hole,
		})
	}

	if tee != "" {
		for i, h := range crs.Holes {
			if h.ID == hole {

				crs.Holes[i].Tees = append(crs.Holes[i].Tees, rnd.Tee{
					ID:  tee,
					Loc: rnd.Loc([]float64{lat, lon}),
				})

				for _, p := range h.Pins {
					crs.Holes[i].Pars = append(crs.Holes[i].Pars, rnd.Par{
						Tee: tee,
						Pin: p.ID,
						Par: 3,
					})
				}

				break
			}
		}
	} else {
		for i, h := range crs.Holes {
			if h.ID == hole {

				crs.Holes[i].Pins = append(crs.Holes[i].Pins, rnd.Pin{
					ID:  pin,
					Loc: rnd.Loc([]float64{lat, lon}),
				})

				for _, t := range h.Tees {
					crs.Holes[i].Pars = append(crs.Holes[i].Pars, rnd.Par{
						Tee: t.ID,
						Pin: pin,
						Par: 3,
					})
				}

				break
			}
		}
	}

	crs.DrawSummary()

}

func newparHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)

	row_s := r.URL.Query()["row"]
	row, _ := strconv.Atoi(row_s[0])
	par_s := r.URL.Query()["par"]
	par, _ := strconv.Atoi(par_s[0])

	idx := 0
	for i, h := range crs.Holes {
		for j, _ := range h.Tees {
			for k, _ := range h.Pins {
				if idx == row {
					crs.Holes[i].Pars[j*len(h.Pins)+k].Par = par
					crs.DrawSummary()
					return
				} else {
					idx++
				}
			}
		}
	}
}

func saveCourseHandler(w http.ResponseWriter, r *http.Request) {
	crs.SaveCourseJSON()
	fmt.Println("Course saved.")
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func main() {

	in := os.Args[1]
	// in := "kit_carson.json"

	crs = rnd.ParseCourseJSON(in)
	crs.DrawSummary()

	// make mapbox representation of course
	http.HandleFunc("/", homeHandler)

	// handle drags
	http.HandleFunc("/movepoint", movepointHandler)

	// handle renames
	http.HandleFunc("/text", textHandler)

	// handle new pars
	http.HandleFunc("/newpar", newparHandler)

	// handle new clicks
	http.HandleFunc("/newPOI", newPOIHandler)

	// handle save button
	http.HandleFunc("/save", saveCourseHandler)

	// serve local files as if they were in /data
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("."))))

	// serve up for local use
	http.ListenAndServe(":8081", nil)

}
