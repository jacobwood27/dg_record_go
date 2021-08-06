package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/tkrajina/gpxgo/gpx"
	"github.com/sgreben/piecewiselinear"
)

type timestamp struct {
	time time.Time
	disc string
}
type timestamps []timestamp

func parseTimestamp(filename string) timestamps {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	var ts []timestamp
	ts_layout := "2006-01-02T15:04:05-07:00"
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		dt, _ := time.Parse(ts_layout, line[0])
		ts = append(ts, timestamp{dt, line[1]})
	}
	return timestamps(ts)
}

type locstamp struct {
	time time.Time
	lat  float64
	lon  float64
}
type locstamps []locstamp

func parseGPX(file string) locstamps {
	t, _ := gpx.ParseFile(file)

	var df []locstamp
	for _, track := range t.Tracks {
		for _, segment := range track.Segments {
			for _, point := range segment.Points {
				df = append(df, locstamp{point.Timestamp, point.Latitude, point.Longitude})
			}
		}
	}
	return locstamps(df)
}

type throwstamp struct {
	num  int
	time time.Time
	disc string
	lat  float64
	lon  float64
}
type throwstamps []throwstamp


func (ts timestamps) timelist() []float64 {
    var list []float64
    for _, t := range ts {
        list = append(list, float64(t.time.UnixNano()))
    }
    return list
}


func interpolateGPX(ts timestamps, locs locstamps) throwstamps {
	var tls []throwstamp

	f := piecewiselinear.Function{
		X:ts.
		Y:ts.timelist()
	}

	return tls
}

func main() {
	ts := parseTimestamp("timestamp.csv")

	// gpx := parseGPX("recording.gpx")

	// Want to record the timezone that was used
	tz, off := ts[0].time.Zone()
	fmt.Println(tz, off)

}
