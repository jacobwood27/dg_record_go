package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/tkrajina/gpxgo/gpx"
)

type timestamp struct {
	time time.Time
	disc string
}

func parseTimestamp(filename string) []timestamp {
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
	return ts
}

type location struct {
	time time.Time
	lat  float64
	lon  float64
}

func parseGPX(file string) []location {
	t, _ := gpx.ParseFile(file)

	var df []location
	for _, track := range t.Tracks {
		for _, segment := range track.Segments {
			for _, point := range segment.Points {
				df = append(df, location{point.Timestamp, point.Latitude, point.Longitude})
			}
		}
	}
	return df
}

func main() {
	ts := parseTimestamp("/home/woojac/proj/025_dg_go/test/data/timestamp.csv")

	// gpx := parseGPX("/home/woojac/proj/025_dg_go/test/data/recording.gpx")

	// Want to record the timezone that was used
	tz, off := ts[0].time.Zone()
	fmt.Println(tz, off)

	// fmt.Println((*df_ts)[0])
}
