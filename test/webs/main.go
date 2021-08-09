// forms.go
package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type throwstamp struct {
	num  int
	time time.Time
	disc string
	lat  float64
	lon  float64
}
type throwstamps []throwstamp

func parseRoundRaw(filename string) throwstamps {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	var ts []throwstamp
	ts_layout := "2006-01-02 15:04:05 -0700"
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		num, _ := strconv.Atoi(line[0])
		dt, _ := time.Parse(ts_layout, line[1])
		disc := line[2]
		lat, _ := strconv.ParseFloat(line[3], 64)
		lon, _ := strconv.ParseFloat(line[4], 64)

		ts = append(ts, throwstamp{num, dt, disc, lat, lon})
	}
	return throwstamps(ts)
}

type MovingDot struct {
	Num int
	Lat float64
	Lon float64
}

func mapboxHandler(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("templates/mapbox2.html"))
	tmpl.Execute(w, nil)

	// var t PointDrag
	// err := json.NewDecoder(r.Body).Decode(&t)

	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(t)
}

func update_round_moving_point(ts throwstamps, dot MovingDot) {
	for _, t := range ts {
		if t.num == dot.Num {
			t.lat = 0.0
			t.lon = 0.0
			return
		}
	}
}

func (ts throwstamps) WriteCSV() {
	f, err := os.Create("round_raw.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	if err != nil {
		log.Fatalln("failed to open file", err)
	}

	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"num", "time", "disc", "lat", "lon"})

	for _, t := range ts {
		w.Write([]string{
			fmt.Sprintf("%d", t.num),
			t.time.String(),
			t.disc,
			fmt.Sprintf("%f", t.lat),
			fmt.Sprintf("%f", t.lon),
		})
	}

}

func movepointHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL)

	num_s := r.URL.Query()["num"]
	num, _ := strconv.Atoi(num_s[0])
	lat_s := r.URL.Query()["lat"]
	lat, _ := strconv.ParseFloat(lat_s[0], 64)
	lon_s := r.URL.Query()["lon"]
	lon, _ := strconv.ParseFloat(lon_s[0], 64)

	dot := MovingDot{num, lat, lon}

	os.Chdir("static")

	ts := parseRoundRaw("round_raw.csv")
	update_round_moving_point(ts, dot)
	ts.WriteCSV()
	// cmnd := exec.Command("", "arg")
	//cmnd.Run() // and wait

	os.Chdir("..")

	fmt.Fprint(w, "All ready bud.")
}

func main() {
	http.HandleFunc("/", mapboxHandler)
	http.HandleFunc("/movepoint", movepointHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.ListenAndServe(":8081", nil)
}
