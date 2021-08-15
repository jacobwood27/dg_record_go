package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

type ScoreDataset struct {
	Label                string
	Data                 []string
	BackgroundColor      string `json:"backgroundColor"`
	BorderColor          string `json:"borderColor"`
	PointRadius          int    `json:"pointRadius"`
	PointColor           string `json:"pointColor"`
	PointStrokeColor     string `json:"pointStrokeColor"`
	PointHighlightFill   string `json:"pointHighlightFill"`
	PointHighlightStroke string `json:"pointHighlightStroke"`
	ShowLine             bool   `json:"showLine"`
	Fill                 bool   `json:"fill"`
}

type Score struct {
	Labels   []string       `json:"labels"`
	Datasets []ScoreDataset `json:"datasets"`
}

type Disc struct {
	ID      string  `json:"id"`
	Image   string  `json:"image"`
	Mold    string  `json:"mold"`
	Plastic string  `json:"plastic"`
	Mass    float64 `json:"mass"`
	Numbers string  ` json:"numbers"`
}

type DashRound struct {
	ID     string `json:"id"`
	Date   string `json:"date"`
	Course string `json:"course"`
	Score  string `json:"score"`
}

type Dash struct {
	Scores Score       `json:"scores"`
	Discs  []Disc      `json:"discs"`
	Rounds []DashRound `json:"rounds"`
}

type myDisc struct {
	MyID    string
	DiscID  string
	Plastic string
	Mass    float64
	Image   string
}

type allDisc struct { // id,brand,mold,type,speed,glide,turn,fade
	DiscID string
	Brand  string
	Mold   string
	Type   string
	Speed  float64
	Glide  float64
	Turn   float64
	Fade   float64
}

func parseAllDiscs() []allDisc {
	adiscs_csv := path.Join(rnd.RootDir(), "..", "data", "discs.csv")
	f, _ := os.Open(adiscs_csv)
	defer f.Close()
	var allD []allDisc
	r := csv.NewReader(f)
	r.Read() // burn header
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}

		speed, _ := strconv.ParseFloat(line[4], 64)
		turn, _ := strconv.ParseFloat(line[5], 64)
		glide, _ := strconv.ParseFloat(line[6], 64)
		fade, _ := strconv.ParseFloat(line[7], 64)
		allD = append(allD, allDisc{
			DiscID: line[0],
			Brand:  line[1],
			Mold:   line[2],
			Type:   line[3],
			Speed:  speed,
			Glide:  glide,
			Turn:   turn,
			Fade:   fade,
		})
	}
	return allD
}

func parseMyDiscs() []myDisc {
	homedir, _ := os.UserHomeDir()
	mydiscsCSV := filepath.Join(homedir, ".discgolf", "discs", "discs.csv")
	f, _ := os.Open(mydiscsCSV)
	defer f.Close()
	var myD []myDisc
	r := csv.NewReader(f)
	r.Read() // burn header
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}

		mass, _ := strconv.ParseFloat(line[3], 64)
		myD = append(myD, myDisc{
			MyID:    line[0],
			DiscID:  line[1],
			Plastic: line[2],
			Mass:    mass,
			Image:   line[4],
		})
	}
	return myD
}

func getDiscs() []Disc {

	var D []Disc
	myD := parseMyDiscs()
	allD := parseAllDiscs()

	for _, d := range myD {
		for _, ad := range allD {
			if d.DiscID == ad.DiscID {

				nums := fmt.Sprintf("%g, %g, %g, %g", ad.Speed, ad.Glide, ad.Turn, ad.Fade)
				D = append(D, Disc{
					ID:      d.MyID,
					Image:   d.Image,
					Mold:    ad.Mold,
					Plastic: d.Plastic,
					Mass:    d.Mass,
					Numbers: nums,
				})
			}
		}
	}

	return D
}

func getScores() Score {
	AR := ParseAllRoundsCSV()

	var L []string
	var ds []string
	var DS []ScoreDataset
	for _, r := range AR {
		L = append(L, r.Date)
		ds = append(ds, strconv.Itoa(r.Score))
	}

	DS = append(DS, ScoreDataset{
		Label:                "Recent Scores",
		Data:                 ds,
		BackgroundColor:      "rgba(60,141,188,0.9)",
		BorderColor:          "rgba(60,141,188,0.8)",
		PointRadius:          5,
		PointColor:           "#3b8bba",
		PointStrokeColor:     "rgba(60,141,188,1)",
		PointHighlightFill:   "#fff",
		PointHighlightStroke: "rgba(60,141,188,1)",
		ShowLine:             true,
		Fill:                 false,
	})

	return Score{
		Labels:   L,
		Datasets: DS,
	}

}

func getRounds() []DashRound {
	var D []DashRound

	AR := ParseAllRoundsCSV()
	for _, r := range AR {
		score := strconv.Itoa(r.Score)
		D = append(D, DashRound{
			ID:     r.Round,
			Date:   r.Date,
			Course: r.Course,
			Score:  score,
		})
	}

	return D
}

func MakeDash() {
	homedir, _ := os.UserHomeDir()
	dashFile := filepath.Join(homedir, ".discgolf", "stats", "dash.json")

	scores := getScores()
	discs := getDiscs()
	rounds := getRounds()

	D := Dash{
		Scores: scores,
		Discs:  discs,
		Rounds: rounds,
	}

	file, _ := json.MarshalIndent(D, "", "	")
	_ = ioutil.WriteFile(dashFile, file, 0644)

}
