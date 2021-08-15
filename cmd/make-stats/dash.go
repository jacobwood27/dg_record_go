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

type LineDataset struct {
	Label                string `json:"label"`
	Data                 []int  `json:"data"`
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

type LinePlot struct {
	Labels   []string      `json:"labels"`
	Datasets []LineDataset `json:"datasets"`
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
	Scores LinePlot    `json:"scores"`
	Discs  []Disc      `json:"discs"`
	Rounds []DashRound `json:"rounds"`
	Putts  LinePlot    `json:"putts"`
	Drives LinePlot    `json:"drives"`
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

func getScores() LinePlot {
	AR := ParseAllRoundsCSV()

	var L []string
	var ds []int
	var DS []LineDataset
	for _, r := range AR {
		L = append(L, r.Date)
		ds = append(ds, r.Score)
	}

	DS = append(DS, LineDataset{
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

	return LinePlot{
		Labels:   L,
		Datasets: DS,
	}

}

func getPutts() LinePlot {
	AT := ParseAllThrowsCSV()

	var L []string
	var ds_10_tries []int
	var ds_10_makes []int
	var ds_20_tries []int
	var ds_20_makes []int
	var ds_33_tries []int
	var ds_33_makes []int
	var ds_66_tries []int
	var ds_66_makes []int

	cur_round := ""
	i := -1
	for _, r := range AT {
		if cur_round != r.Round {
			cur_round = r.Round
			L = append(L, r.Date)
			ds_10_tries = append(ds_10_tries, 0)
			ds_10_makes = append(ds_10_makes, 0)
			ds_20_tries = append(ds_20_tries, 0)
			ds_20_makes = append(ds_20_makes, 0)
			ds_33_tries = append(ds_33_tries, 0)
			ds_33_makes = append(ds_33_makes, 0)
			ds_66_tries = append(ds_66_tries, 0)
			ds_66_makes = append(ds_66_makes, 0)
			i++
		}

		if r.DistPin < 10 {
			ds_10_tries[i]++
			if r.ResThrow == "MAKE" || r.ResThrow == "ACE" {
				ds_10_makes[i]++
			}
		} else if r.DistPin < 20 {
			ds_20_tries[i]++
			if r.ResThrow == "MAKE" || r.ResThrow == "ACE" {
				ds_20_makes[i]++
			}
		} else if r.DistPin < 33 {
			ds_33_tries[i]++
			if r.ResThrow == "MAKE" || r.ResThrow == "ACE" {
				ds_33_makes[i]++
			}
		} else if r.DistPin < 66 {
			ds_66_tries[i]++
			if r.ResThrow == "MAKE" || r.ResThrow == "ACE" {
				ds_66_makes[i]++
			}
		}
	}

	var DS []LineDataset

	var ds_10 []int
	for i, _ := range ds_10_tries {
		ds_10 = append(ds_10, (100*ds_10_makes[i])/ds_10_tries[i])
	}
	DS = append(DS, LineDataset{
		Label:                "0 - 10 ft",
		Data:                 ds_10,
		BackgroundColor:      "rgba(167, 36, 193, 1)",
		BorderColor:          "rgba(167, 36, 193, 1)",
		PointRadius:          5,
		PointColor:           "#3b8bba",
		PointStrokeColor:     "rgba(167, 36, 193, 1)",
		PointHighlightFill:   "#fff",
		PointHighlightStroke: "rgba(167, 36, 193, 1)",
		ShowLine:             true,
		Fill:                 false,
	})

	var ds_20 []int
	for i, _ := range ds_20_makes {
		ds_20 = append(ds_20, (100*ds_20_makes[i])/ds_20_tries[i])
	}
	DS = append(DS, LineDataset{
		Label:                "10 - 20 ft",
		Data:                 ds_20,
		BackgroundColor:      "rgba(0, 188, 212, 1)",
		BorderColor:          "rgba(0, 188, 212, 1)",
		PointRadius:          5,
		PointColor:           "#3b8bba",
		PointStrokeColor:     "rgba(0, 188, 212, 1)",
		PointHighlightFill:   "#fff",
		PointHighlightStroke: "rgba(0, 188, 212, 1)",
		ShowLine:             true,
		Fill:                 false,
	})

	var ds_33 []int
	for i, _ := range ds_33_makes {
		ds_33 = append(ds_33, (100*ds_33_makes[i])/ds_33_tries[i])
	}
	DS = append(DS, LineDataset{
		Label:                "20 - 33 ft",
		Data:                 ds_33,
		BackgroundColor:      "rgba(214, 220, 57, 1)",
		BorderColor:          "rgba(214, 220, 57, 1)",
		PointRadius:          5,
		PointColor:           "#3b8bba",
		PointStrokeColor:     "rgba(214, 220, 57, 1)",
		PointHighlightFill:   "#fff",
		PointHighlightStroke: "rgba(214, 220, 57, 1)",
		ShowLine:             true,
		Fill:                 false,
	})

	var ds_66 []int
	for i, _ := range ds_66_makes {
		ds_66 = append(ds_66, 100*(ds_66_makes[i])/ds_66_tries[i])
	}
	DS = append(DS, LineDataset{
		Label:                "33 - 66 ft",
		Data:                 ds_66,
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

	return LinePlot{
		Labels:   L,
		Datasets: DS,
	}

}

func getDrives() LinePlot {
	AT := ParseAllThrowsCSV()

	var L []string
	var ds []int

	cur_round := ""
	i := -1
	for _, r := range AT {
		if cur_round != r.Round {
			cur_round = r.Round
			L = append(L, r.Date)
			ds = append(ds, 0)
			i++
		}

		if int(r.Dist) > ds[i] {
			ds[i] = int(r.Dist)
		}
	}

	var DS []LineDataset
	DS = append(DS, LineDataset{
		Label:                "Long Drive",
		Data:                 ds,
		BackgroundColor:      "rgba(167, 36, 193, 1)",
		BorderColor:          "rgba(167, 36, 193, 1)",
		PointRadius:          5,
		PointColor:           "#3b8bba",
		PointStrokeColor:     "rgba(167, 36, 193, 1)",
		PointHighlightFill:   "#fff",
		PointHighlightStroke: "rgba(167, 36, 193, 1)",
		ShowLine:             true,
		Fill:                 false,
	})

	return LinePlot{
		Labels:   L,
		Datasets: DS,
	}

}

func getRounds() []DashRound {
	var D []DashRound

	AR := ParseAllRoundsCSV()
	// for i, r := range AR {
	for i := len(AR) - 1; i >= 0; i-- {
		r := AR[i]
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

	D := Dash{
		Scores: getScores(),
		Discs:  getDiscs(),
		Rounds: getRounds(),
		Putts:  getPutts(),
		Drives: getDrives(),
	}

	file, _ := json.MarshalIndent(D, "", "	")
	_ = ioutil.WriteFile(dashFile, file, 0644)

}
