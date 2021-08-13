package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/sgreben/piecewiselinear"
	"github.com/tkrajina/gpxgo/gpx"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

type timestamp struct {
	time time.Time
	disc string
}
type timestamps []timestamp

func parseTimestampFile(filename string) timestamps {
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

func parseGPXFile(file string) locstamps {
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

func interpolateGPX(ts timestamps, locs locstamps) rnd.Stamps {
	var tls []rnd.Stamp

	f_lat := piecewiselinear.Function{
		X: make([]float64, 0),
		Y: make([]float64, 0),
	}
	f_lon := piecewiselinear.Function{
		X: make([]float64, 0),
		Y: make([]float64, 0),
	}
	for _, l := range locs {
		f_lat.X = append(f_lat.X, float64(l.time.UnixNano()))
		f_lat.Y = append(f_lat.Y, l.lat)
		f_lon.X = append(f_lon.X, float64(l.time.UnixNano()))
		f_lon.Y = append(f_lon.Y, l.lon)
	}

	for _, t := range ts {
		tls = append(tls, rnd.Stamp{
			Loc: rnd.Loc{
				f_lat.At(float64(t.time.UnixNano())),
				f_lon.At(float64(t.time.UnixNano())),
			},
			Disc: t.disc,
		})
	}

	return rnd.Stamps(tls)
}

func main() {
	ts := parseTimestampFile("timestamp.csv")

	gpx := parseGPXFile("recording.gpx")

	igpx := interpolateGPX(ts, gpx)

	fID := ts[0].time.Format("2006-01-02-15-04-05")
	igpx.WriteRoundRawCSV(fID)
}
