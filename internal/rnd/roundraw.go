package rnd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

type Throwstamp struct {
	Num  int
	Time time.Time
	Disc string
	Lat  float64
	Lon  float64
}
type Throwstamps []Throwstamp

func (t Throwstamp) Loc() Loc {
	return Loc{t.Lat, t.Lon}
}

func (ts Throwstamps) WriteCSV() {
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
			fmt.Sprintf("%d", t.Num),
			t.Time.String(),
			t.Disc,
			fmt.Sprintf("%f", t.Lat),
			fmt.Sprintf("%f", t.Lon),
		})
	}

}

func ReadRoundRawCSV(filename string) Throwstamps {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	_, err = r.Read() // Burn the header line
	if err != nil {
		log.Fatal(err)
	}

	var ts []Throwstamp
	ts_layout := "2006-01-02 15:04:05 -0700 MST"
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		num, _ := strconv.Atoi(line[0])
		t, _ := time.Parse(ts_layout, line[1])
		lat, _ := strconv.ParseFloat(line[3], 64)
		lon, _ := strconv.ParseFloat(line[4], 64)
		ts = append(ts, Throwstamp{
			Num:  num,
			Time: t,
			Disc: line[2],
			Lat:  lat,
			Lon:  lon})
	}
	return Throwstamps(ts)

}
