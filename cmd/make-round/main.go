package main

import (
	"path/filepath"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

func main() {

	matches, _ := filepath.Glob("./*_raw.csv")

	for _, m := range matches {
		fID := m[:len(m)-8]
		round, _ := rnd.ParseRoundRaw(fID)
		round.WriteRoundJSON()
		round.MakeRoundVis()
	}
}
