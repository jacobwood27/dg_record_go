package rnd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

type Stamp struct {
	Loc  Loc    `json:"loc"`
	Disc string `json:"disc"`
}
type Stamps []Stamp

func (ts Stamps) WriteRoundRawCSV(fID string) {
	f, err := os.Create(fID + "_raw.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"lat", "lon", "disc"})

	for _, t := range ts {
		w.Write([]string{
			fmt.Sprintf("%f", t.Loc[0]),
			fmt.Sprintf("%f", t.Loc[1]),
			t.Disc,
		})
	}
}

func ReadRoundRawCSV(fileID string) Stamps {
	f, err := os.Open(fileID + "_raw.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	_, err = r.Read() // Burn the header line
	if err != nil {
		log.Fatal(err)
	}

	var ts []Stamp
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		lat, _ := strconv.ParseFloat(line[0], 64)
		lon, _ := strconv.ParseFloat(line[1], 64)
		disc := line[2]
		ts = append(ts, Stamp{
			Loc:  Loc{lat, lon},
			Disc: disc})
	}

	return Stamps(ts)
}
