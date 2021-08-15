package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jacobwood27/go-dg-record/internal/rnd"
)

func rndFiles() []string {
	homedir, _ := os.UserHomeDir()
	rnddir := filepath.Join(homedir, ".discgolf", "rounds")

	files, err := ioutil.ReadDir(rnddir)
	if err != nil {
		log.Fatal(err)
	}

	var S []string
	for _, f := range files {
		S = append(S, filepath.Join(rnddir, f.Name(), f.Name()+".csv"))
	}

	return S
}

func main() {

	rfiles := rndFiles()
	var rnds []rnd.Round
	for _, rfile := range rfiles {
		rd := rnd.ReadRoundCSV(rfile)
		rnds = append(rnds, rd)
	}

	MakeAllThrowsCSV(rnds)
	MakeAllHolesCSV(rnds)
	MakeAllRoundsCSV(rnds)
	MakeDash()
}
